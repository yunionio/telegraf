//go:generate ../../../tools/readme_config_includer/generator
package diskio

import (
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/disk"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/common/psutil"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

var (
	varRegex = regexp.MustCompile(`\$(?:\w+|\{\w+\})`)
)

type DiskIO struct {
	Devices          []string        `toml:"devices"`
	DeviceTags       []string        `toml:"device_tags"`
	NameTemplates    []string        `toml:"name_templates"`
	Excludes         string          `toml:"excludes"`
	SkipSerialNumber bool            `toml:"skip_serial_number"`
	Log              telegraf.Logger `toml:"-"`

	ps                psutil.PS
	infoCache         map[string]diskInfoCache
	deviceFilter      filter.Filter
	warnDiskName      map[string]bool
	warnDiskTags      map[string]bool
	lastIOCounterStat map[string]disk.IOCountersStat
	lastCollectTime   time.Time
}

func (*DiskIO) SampleConfig() string {
	return sampleConfig
}

func (d *DiskIO) Init() error {
	for _, device := range d.Devices {
		if hasMeta(device) {
			deviceFilter, err := filter.Compile(d.Devices)
			if err != nil {
				return fmt.Errorf("error compiling device pattern: %w", err)
			}
			d.deviceFilter = deviceFilter
		}
	}

	d.infoCache = make(map[string]diskInfoCache)
	d.warnDiskName = make(map[string]bool)
	d.warnDiskTags = make(map[string]bool)
	d.lastIOCounterStat = make(map[string]disk.IOCountersStat)

	return nil
}

func (d *DiskIO) Gather(acc telegraf.Accumulator) error {
	var devices []string
	if d.deviceFilter == nil {
		for _, dev := range d.Devices {
			devices = append(devices, resolveName(dev))
		}
	}

	diskio, err := d.ps.DiskIO(devices)
	if err != nil {
		return fmt.Errorf("error getting disk io info: %w", err)
	}
	collectTime := time.Now()

	var excludeReg *regexp.Regexp
	if len(d.Excludes) > 0 {
		excludeReg = regexp.MustCompile(d.Excludes)
	}
	for k, io := range diskio {
		if excludeReg != nil && excludeReg.MatchString(io.Name) {
			continue
		}

		match := false
		if d.deviceFilter != nil && d.deviceFilter.Match(io.Name) {
			match = true
		}

		tags := make(map[string]string)
		var devLinks []string
		tags["name"], devLinks = d.diskName(io.Name)

		if wwid := getDeviceWWID(io.Name); wwid != "" {
			tags["wwid"] = wwid
		}

		if d.deviceFilter != nil && !match {
			for _, devLink := range devLinks {
				if d.deviceFilter.Match(devLink) {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}

		for t, v := range d.diskTags(io.Name) {
			tags[t] = v
		}

		if !d.SkipSerialNumber {
			if len(io.SerialNumber) != 0 {
				tags["serial"] = io.SerialNumber
			} else {
				tags["serial"] = "unknown"
			}
		}

		fields := map[string]interface{}{
			"reads":            io.ReadCount,
			"writes":           io.WriteCount,
			"iobytes":          io.ReadBytes + io.WriteBytes,
			"read_bytes":       io.ReadBytes,
			"write_bytes":      io.WriteBytes,
			"read_time":        io.ReadTime,
			"write_time":       io.WriteTime,
			"io_time":          io.IoTime,
			"weighted_io_time": io.WeightedIO,
			"iops_in_progress": io.IopsInProgress,
			"iocount":          io.ReadCount + io.WriteCount,
			"merged_reads":     io.MergedReadCount,
			"merged_writes":    io.MergedWriteCount,
			"merged_iocount":   io.MergedReadCount + io.MergedWriteCount,
		}
		if lastValue, exists := d.lastIOCounterStat[k]; exists {
			// Check for wrap around
			wrap := io.ReadCount < lastValue.ReadCount || io.WriteCount < lastValue.WriteCount
			wrap = wrap || io.ReadTime < lastValue.ReadTime || io.WriteTime < lastValue.WriteTime
			wrap = wrap || io.IoTime < lastValue.IoTime

			if !wrap {
				deltaRWCount := float64(io.ReadCount-lastValue.ReadCount) + float64(io.WriteCount-lastValue.WriteCount)
				deltaRWTime := float64(io.ReadTime-lastValue.ReadTime) + float64(io.WriteTime-lastValue.WriteTime)
				deltaIOTime := float64(io.IoTime - lastValue.IoTime)

				timeDelta := float64(collectTime.Sub(d.lastCollectTime).Seconds())

				readIo := io.ReadCount - lastValue.ReadCount
				writeIo := io.WriteCount - lastValue.WriteCount
				fields["iops"] = float64(readIo+writeIo) / timeDelta
				readBytes := io.ReadBytes - lastValue.ReadBytes
				fields["read_iops"] = float64(readIo) / timeDelta
				writeBytes := io.WriteBytes - lastValue.WriteBytes
				fields["write_iops"] = float64(writeIo) / timeDelta
				fields["bps"] = float64(readBytes+writeBytes) / timeDelta
				fields["read_bps"] = float64(readBytes) / timeDelta
				fields["write_bps"] = float64(writeBytes) / timeDelta
				readTime := io.ReadTime - lastValue.ReadTime
				writeTime := io.WriteTime - lastValue.WriteTime

				ioTime := io.IoTime - lastValue.IoTime
				weightedIoTime := io.WeightedIO - lastValue.WeightedIO
				readAwait := 0.0
				if readIo > 0 {
					readAwait = float64(readTime) / float64(readIo)
				}
				writeAwait := 0.0
				if writeIo > 0 {
					writeAwait = float64(writeTime) / float64(writeIo)
				}
				ioAwait := 0.0
				if readIo+writeIo > 0 {
					ioAwait = float64(readTime+writeTime) / float64(readIo+writeIo)
				}
				fields["read_awit"] = readAwait
				fields["write_awit"] = writeAwait
				fields["await"] = ioAwait
				fields["ioutil"] = float64(ioTime*100) / timeDelta / 1000.0
				fields["avgqu_sz"] = float64(weightedIoTime) / timeDelta / 1000.0

				if deltaRWCount > 0 {
					fields["io_await"] = deltaRWTime / deltaRWCount
					fields["io_svctm"] = deltaIOTime / deltaRWCount
				}

				itv := float64(collectTime.Sub(d.lastCollectTime).Milliseconds())
				if itv > 0 {
					fields["io_util"] = 100 * deltaIOTime / itv
				}
			}
		}
		acc.AddCounter("diskio", fields, tags)
	}
	d.lastCollectTime = collectTime
	d.lastIOCounterStat = diskio
	return nil
}

// hasMeta reports whether s contains any special glob characters.
func hasMeta(s string) bool {
	return strings.ContainsAny(s, "*?[")
}

func (d *DiskIO) diskName(devName string) (string, []string) {
	di, err := d.diskInfo(devName)
	devLinks := strings.Split(di["DEVLINKS"], " ")
	for i, devLink := range devLinks {
		devLinks[i] = strings.TrimPrefix(devLink, "/dev/")
	}
	// Return error after attempting to process some of the devlinks.
	// These could exist if we got further along the diskInfo call.
	if err != nil {
		if ok := d.warnDiskName[devName]; !ok {
			d.warnDiskName[devName] = true
			d.Log.Warnf("Unable to gather disk name for %q: %s", devName, err)
		}
		return devName, devLinks
	}

	if len(d.NameTemplates) == 0 {
		return devName, devLinks
	}

	for _, nt := range d.NameTemplates {
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
			return name, devLinks
		}
	}

	return devName, devLinks
}

func (d *DiskIO) diskTags(devName string) map[string]string {
	if len(d.DeviceTags) == 0 {
		return nil
	}

	di, err := d.diskInfo(devName)
	if err != nil {
		if ok := d.warnDiskTags[devName]; !ok {
			d.warnDiskTags[devName] = true
			d.Log.Warnf("Unable to gather disk tags for %q: %s", devName, err)
		}
		return nil
	}

	tags := make(map[string]string, len(d.DeviceTags))
	for _, dt := range d.DeviceTags {
		if v, ok := di[dt]; ok {
			tags[dt] = v
		}
	}

	return tags
}

func init() {
	ps := psutil.NewSystemPS()
	inputs.Add("diskio", func() telegraf.Input {
		return &DiskIO{ps: ps, SkipSerialNumber: true}
	})
}
