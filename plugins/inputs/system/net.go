package system

import (
	"fmt"
	"net"
	"strings"
	"time"
	"io/ioutil"
	"strconv"

	psnet "github.com/shirou/gopsutil/net"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type NetIOStats struct {
	filter filter.Filter
	ps     PS

	skipChecks bool
	Interfaces []string
	Speed int

	lastTime time.Time
	lastStats map[string]psnet.IOCountersStat
}

func (_ *NetIOStats) Description() string {
	return "Read metrics about network interface usage"
}

var netSampleConfig = `
  ## By default, telegraf gathers stats from any up interface (excluding loopback)
  ## Setting interfaces will tell it to gather these explicit interfaces,
  ## regardless of status.
  ##
  # interfaces = ["eth0"]
  # speed = 1000
`

func (_ *NetIOStats) SampleConfig() string {
	return netSampleConfig
}

func (s *NetIOStats) interfaceSpeed(name string) int {
	path := fmt.Sprintf("/sys/class/net/%s/speed", name)
	content, err := ioutil.ReadFile(path)
	speed := 0
	if err == nil {
		speed, err = strconv.Atoi(string(content))
		if err != nil {
			speed = 0
		}
	}
	if speed == 0 {
		speed = s.Speed
	}
	return speed
}

func (s *NetIOStats) Gather(acc telegraf.Accumulator) error {
	netio, err := s.ps.NetIO()
	if err != nil {
		return fmt.Errorf("error getting net io info: %s", err)
	}

	if s.filter == nil {
		if s.filter, err = filter.Compile(s.Interfaces); err != nil {
			return fmt.Errorf("error compiling filter: %s", err)
		}
	}

	curr := time.Now()
	timeDelta := curr.Sub(s.lastTime).Seconds()

	for _, io := range netio {
		if len(s.Interfaces) != 0 {
			var found bool

			if s.filter.Match(io.Name) {
				found = true
			}

			if !found {
				continue
			}
		} else if !s.skipChecks {
			iface, err := net.InterfaceByName(io.Name)
			if err != nil {
				continue
			}

			if iface.Flags&net.FlagLoopback == net.FlagLoopback {
				continue
			}

			if iface.Flags&net.FlagUp == 0 {
				continue
			}
		}

		tags := map[string]string{
			"interface": io.Name,
		}

		fields := map[string]interface{}{
			"bytes_sent":   io.BytesSent,
			"bytes_recv":   io.BytesRecv,
			"packets_sent": io.PacketsSent,
			"packets_recv": io.PacketsRecv,
			"err_in":       io.Errin,
			"err_out":      io.Errout,
			"drop_in":      io.Dropin,
			"drop_out":     io.Dropout,
		}
		acc.AddCounter("net", fields, tags, curr)

		if len(s.lastStats) == 0 {
			continue
		}

		last, ok := s.lastStats[io.Name]
		if !ok {
			continue
		}

		bps_sent := float64(io.BytesSent - last.BytesSent)*8.0/timeDelta
		bps_recv := float64(io.BytesRecv - last.BytesRecv)*8.0/timeDelta
		fields2 := map[string]interface{}{
			"bps_sent": bps_sent,
			"bps_recv": bps_recv,
			"pps_sent": float64(io.PacketsSent - last.PacketsSent)/timeDelta,
			"pps_recv": float64(io.PacketsRecv - last.PacketsRecv)/timeDelta,
			"pps_err_in": float64(io.Errin - last.Errin)/timeDelta,
			"pps_err_out": float64(io.Errout - last.Errout)/timeDelta,
			"pps_drop_in": float64(io.Dropin - last.Dropin)/timeDelta,
			"pps_drop_out": float64(io.Dropout - last.Dropout)/timeDelta,
		}
		speed := s.interfaceSpeed(io.Name)
		fields2["speed"] = speed
		if speed > 0 {
			fields2["if_in_percent"] = bps_recv/float64(speed)/10000.0
			fields2["if_out_percent"] = bps_sent/float64(speed)/10000.0
		}
		acc.AddGauge("net", fields2, tags, curr)
	}

	s.lastStats = make(map[string]psnet.IOCountersStat)
	for _, io := range netio {
		s.lastStats[io.Name] = io
	}
	s.lastTime = curr

	// Get system wide stats for different network protocols
	// (ignore these stats if the call fails)
	netprotos, _ := s.ps.NetProto()
	fields := make(map[string]interface{})
	for _, proto := range netprotos {
		for stat, value := range proto.Stats {
			name := fmt.Sprintf("%s_%s", strings.ToLower(proto.Protocol),
				strings.ToLower(stat))
			fields[name] = value
		}
	}
	tags := map[string]string{
		"interface": "all",
	}
	acc.AddFields("net", fields, tags, curr)

	return nil
}

func init() {
	inputs.Add("net", func() telegraf.Input {
		return &NetIOStats{ps: newSystemPS()}
	})
}
