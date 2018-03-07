package system

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/shirou/gopsutil/disk"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type DiskStats struct {
	ps PS

	// Legacy support
	Mountpoints []string

	MountPoints []string
	IgnoreFS    []string `toml:"ignore_fs"`
}

func (_ *DiskStats) Description() string {
	return "Read metrics about disk usage by mount point"
}

var diskSampleConfig = `
  ## By default, telegraf gather stats for all mountpoints.
  ## Setting mountpoints will restrict the stats to the specified mountpoints.
  # mount_points = ["/"]

  ## Ignore some mountpoints by filesystem type. For example (dev)tmpfs (usually
  ## present on /run, /var/run, /dev/shm or /dev).
  ignore_fs = ["tmpfs", "devtmpfs", "devfs"]
`

func (_ *DiskStats) SampleConfig() string {
	return diskSampleConfig
}

func (s *DiskStats) Gather(acc telegraf.Accumulator) error {
	// Legacy support:
	if len(s.Mountpoints) != 0 {
		s.MountPoints = s.Mountpoints
	}

	disks, partitions, err := s.ps.DiskUsage(s.MountPoints, s.IgnoreFS)
	if err != nil {
		return fmt.Errorf("error getting disk usage info: %s", err)
	}

	for i, du := range disks {
		if du.Total == 0 {
			// Skip dummy filesystem (procfs, cgroupfs, ...)
			continue
		}
		mountOpts := parseOptions(partitions[i].Opts)
		mode := mountOpts.Mode()
		tags := map[string]string{
			"path":   du.Path,
			"device": strings.Replace(partitions[i].Device, "/dev/", "", -1),
			"fstype": du.Fstype,
			"mode":   mode,
		}
		var used_percent float64
		if du.Used+du.Free > 0 {
			used_percent = float64(du.Used) /
				(float64(du.Used) + float64(du.Free)) * 100
		}
		ro := 0
		if mode == "ro" {
			ro = 1
		}
		var inodesUsedPercent float64
		if du.InodesFree + du.InodesUsed > 0 {
			inodesUsedPercent = float64(du.InodesUsed) /
				(float64(du.InodesFree) + float64(du.InodesUsed)) * 100
		}
		fields := map[string]interface{}{
			"total":        du.Total,
			"free":         du.Free,
			"used":         du.Used,
			"used_percent": used_percent,
			"inodes_total": du.InodesTotal,
			"inodes_free":  du.InodesFree,
			"inodes_used":  du.InodesUsed,
			"inodes_used_percent": inodesUsedPercent,
			"read_only":    ro,
		}
		acc.AddGauge("disk", fields, tags)
	}

	return nil
}

type DiskIOStats struct {
	ps PS

	Devices          []string
	DeviceTags       []string
	NameTemplates    []string
	SkipSerialNumber bool

	infoCache map[string]diskInfoCache

	lastStats map[string]disk.IOCountersStat
	lastTime  time.Time
}

func (_ *DiskIOStats) Description() string {
	return "Read metrics about disk IO by device"
}

var diskIoSampleConfig = `
  ## By default, telegraf will gather stats for all devices including
  ## disk partitions.
  ## Setting devices will restrict the stats to the specified devices.
  # devices = ["sda", "sdb"]
  ## Uncomment the following line if you need disk serial numbers.
  # skip_serial_number = false
  #
  ## On systems which support it, device metadata can be added in the form of
  ## tags.
  ## Currently only Linux is supported via udev properties. You can view
  ## available properties for a device by running:
  ## 'udevadm info -q property -n /dev/sda'
  # device_tags = ["ID_FS_TYPE", "ID_FS_USAGE"]
  #
  ## Using the same metadata source as device_tags, you can also customize the
  ## name of the device via templates.
  ## The 'name_templates' parameter is a list of templates to try and apply to
  ## the device. The template may contain variables in the form of '$PROPERTY' or
  ## '${PROPERTY}'. The first template which does not contain any variables not
  ## present for the device is used as the device name tag.
  ## The typical use case is for LVM volumes, to get the VG/LV name instead of
  ## the near-meaningless DM-0 name.
  # name_templates = ["$ID_FS_LABEL","$DM_VG_NAME/$DM_LV_NAME"]
`

func (_ *DiskIOStats) SampleConfig() string {
	return diskIoSampleConfig
}

func (s *DiskIOStats) Gather(acc telegraf.Accumulator) error {
	diskio, err := s.ps.DiskIO(s.Devices)
	if err != nil {
		return fmt.Errorf("error getting disk io info: %s", err)
	}

	curr := time.Now()
	timeDelta := curr.Sub(s.lastTime).Seconds()

	for _, io := range diskio {
		tags := map[string]string{}
		tags["name"] = s.diskName(io.Name)
		for t, v := range s.diskTags(io.Name) {
			tags[t] = v
		}
		if !s.SkipSerialNumber {
			if len(io.SerialNumber) != 0 {
				tags["serial"] = io.SerialNumber
			} else {
				tags["serial"] = "unknown"
			}
		}

		fields := map[string]interface{}{
			"reads":            io.ReadCount,
			"writes":           io.WriteCount,
			"iocount":          io.ReadCount + io.WriteCount,
			"merged_reads":     io.MergedReadCount,
			"merged_writes":    io.MergedWriteCount,
			"merged_iocount":   io.MergedReadCount + io.MergedWriteCount,
			"read_bytes":       io.ReadBytes,
			"write_bytes":      io.WriteBytes,
			"iobytes":          io.ReadBytes + io.WriteBytes,
			"read_time":        io.ReadTime, // ms
			"write_time":       io.WriteTime, // ms
			"io_time":          io.IoTime, // ms
			"weighted_io_time": io.WeightedIO, // ms
			"iops_in_progress": io.IopsInProgress,
		}
		acc.AddCounter("diskio", fields, tags, curr)

		if len(s.lastStats) == 0 {
			// If it's the 1st gather, can't get CPU Usage stats yet
			continue
		}

		last, ok := s.lastStats[io.Name]
		if !ok {
			continue
		}

		readIo := io.ReadCount - last.ReadCount
		writeIo := io.WriteCount - last.WriteCount
		readBytes := io.ReadBytes - last.ReadBytes
		writeBytes := io.WriteBytes - last.WriteBytes
		readTime := io.ReadTime - last.ReadTime
		writeTime := io.WriteTime - last.WriteTime
		ioTime := io.IoTime - last.IoTime
		weightedIoTime := io.WeightedIO - last.WeightedIO
		readAwait := 0.0
		if readIo > 0 {
			readAwait = float64(readTime)/float64(readIo)
		}
		writeAwait := 0.0
		if writeIo > 0 {
			writeAwait = float64(writeTime)/float64(writeIo)
		}
		ioAwait := 0.0
		if readIo + writeIo > 0 {
			ioAwait = float64(readTime + writeTime)/float64(readIo + writeIo)
		}

		fields2 := map[string]interface{}{
			"iops": float64(readIo + writeIo)/timeDelta,
			"read_iops": float64(readIo)/timeDelta,
			"write_iops": float64(writeIo)/timeDelta,
			"read_bps": float64(readBytes)/timeDelta,
			"write_bps": float64(writeBytes)/timeDelta,
			"read_await": readAwait,
			"write_await": writeAwait,
			"await": ioAwait,
			"ioutil": float64(ioTime*100)/timeDelta/1000.0,
			"avgqu_sz": float64(weightedIoTime)/timeDelta/1000.0,
		}
		acc.AddGauge("diskio", fields2, tags, curr)
	}

	s.lastStats = make(map[string]disk.IOCountersStat)
	for _, io := range diskio {
		s.lastStats[io.Name] = io
	}
	s.lastTime = curr

	return nil
}

var varRegex = regexp.MustCompile(`\$(?:\w+|\{\w+\})`)

func (s *DiskIOStats) diskName(devName string) string {
	if len(s.NameTemplates) == 0 {
		return devName
	}

	di, err := s.diskInfo(devName)
	if err != nil {
		log.Printf("W! Error gathering disk info: %s", err)
		return devName
	}

	for _, nt := range s.NameTemplates {
		miss := false
		name := varRegex.ReplaceAllStringFunc(nt, func(sub string) string {
			sub = sub[1:] // strip leading '$'
			if sub[0] == '{' {
				sub = sub[1 : len(sub)-1] // strip leading & trailing '{' '}'
			}
			if v, ok := di[sub]; ok {
				return v
			}
			miss = true
			return ""
		})

		if !miss {
			return name
		}
	}

	return devName
}

func (s *DiskIOStats) diskTags(devName string) map[string]string {
	if len(s.DeviceTags) == 0 {
		return nil
	}

	di, err := s.diskInfo(devName)
	if err != nil {
		log.Printf("W! Error gathering disk info: %s", err)
		return nil
	}

	tags := map[string]string{}
	for _, dt := range s.DeviceTags {
		if v, ok := di[dt]; ok {
			tags[dt] = v
		}
	}

	return tags
}

type MountOptions []string

func (opts MountOptions) Mode() string {
	if opts.exists("rw") {
		return "rw"
	} else if opts.exists("ro") {
		return "ro"
	} else {
		return "unknown"
	}
}

func (opts MountOptions) exists(opt string) bool {
	for _, o := range opts {
		if o == opt {
			return true
		}
	}
	return false
}

func parseOptions(opts string) MountOptions {
	return strings.Split(opts, ",")
}

func init() {
	ps := newSystemPS()
	inputs.Add("disk", func() telegraf.Input {
		return &DiskStats{ps: ps}
	})

	inputs.Add("diskio", func() telegraf.Input {
		return &DiskIOStats{ps: ps, SkipSerialNumber: true}
	})
}
