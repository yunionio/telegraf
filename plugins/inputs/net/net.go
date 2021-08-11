package net

import (
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"time"

	psnet "github.com/shirou/gopsutil/net"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/system"
)

type InterfaceProfile struct {
	Name  string
	Alias string
	Speed int
}

type NetIOStats struct {
	filter filter.Filter
	ps     system.PS

	skipChecks          bool
	IgnoreProtocolStats bool
	Interfaces          []string
	InterfaceConf       []InterfaceProfile

	lastTime  time.Time
	lastStats map[string]psnet.IOCountersStat
}

func (n *NetIOStats) Description() string {
	return "Read metrics about network interface usage"
}

func getInterfaceSpeed(name string) int {
	path := fmt.Sprintf("/sys/class/net/%s/speed", name)
	content, err := ioutil.ReadFile(path)
	speed := 0
	if err == nil {
		speed, err = strconv.Atoi(string(content))
		if err != nil {
			speed = 0
		}
	}
	return speed
}

func (s *NetIOStats) getProfile(name string) *InterfaceProfile {
	for _, inf := range s.InterfaceConf {
		if inf.Name == name {
			return &inf
		}
	}
	return nil
}

var netSampleConfig = `
  ## By default, telegraf gathers stats from any up interface (excluding loopback)
  ## Setting interfaces will tell it to gather these explicit interfaces,
  ## regardless of status.
  ##
  # interfaces = ["eth0"]
  ##
  ## On linux systems telegraf also collects protocol stats.
  ## Setting ignore_protocol_stats to true will skip reporting of protocol metrics.
  ##
  # ignore_protocol_stats = false
  ##
`

func (n *NetIOStats) SampleConfig() string {
	return netSampleConfig
}

func (n *NetIOStats) Gather(acc telegraf.Accumulator) error {
	netio, err := n.ps.NetIO()
	if err != nil {
		return fmt.Errorf("error getting net io info: %s", err)
	}

	if n.filter == nil {
		if n.filter, err = filter.Compile(n.Interfaces); err != nil {
			return fmt.Errorf("error compiling filter: %s", err)
		}
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return fmt.Errorf("error getting list of interfaces: %s", err)
	}
	interfacesByName := map[string]net.Interface{}
	for _, iface := range interfaces {
		interfacesByName[iface.Name] = iface
	}

	curr := time.Now()
	timeDelta := curr.Sub(n.lastTime).Seconds()

	for _, io := range netio {
		if len(n.Interfaces) != 0 {
			var found bool

			if n.filter.Match(io.Name) {
				found = true
			}

			if !found {
				continue
			}
		} else if !n.skipChecks {
			iface, ok := interfacesByName[io.Name]
			if !ok {
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

		prof := n.getProfile(io.Name)
		if prof != nil && len(prof.Alias) > 0 {
			tags["alias"] = prof.Alias
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

		if len(n.lastStats) == 0 {
			continue
		}

		last, ok := n.lastStats[io.Name]
		if !ok {
			continue
		}

		bpsSent := float64(io.BytesSent-last.BytesSent) * 8.0 / timeDelta
		bpsRecv := float64(io.BytesRecv-last.BytesRecv) * 8.0 / timeDelta

		fields2 := map[string]interface{}{
			"bps_sent":     bpsSent,
			"bps_recv":     bpsRecv,
			"pps_sent":     float64(io.PacketsSent-last.PacketsSent) / timeDelta,
			"pps_recv":     float64(io.PacketsRecv-last.PacketsRecv) / timeDelta,
			"pps_err_in":   float64(io.Errin-last.Errin) / timeDelta,
			"pps_err_out":  float64(io.Errout-last.Errout) / timeDelta,
			"pps_drop_in":  float64(io.Dropin-last.Dropin) / timeDelta,
			"pps_drop_out": float64(io.Dropout-last.Dropout) / timeDelta,
		}
		speed := 0
		if prof != nil && prof.Speed > 0 {
			speed = prof.Speed
		} else {
			speed = getInterfaceSpeed(io.Name)
		}
		fields2["speed"] = speed
		if speed > 0 {
			fields2["if_in_percent"] = bpsRecv / float64(speed) / 10000.0
			fields2["if_out_percent"] = bpsSent / float64(speed) / 10000.0
		}
		acc.AddGauge("net", fields2, tags, curr)
	}

	n.lastStats = make(map[string]psnet.IOCountersStat)
	for _, io := range netio {
		n.lastStats[io.Name] = io
	}
	n.lastTime = curr

	// Get system wide stats for different network protocols
	// (ignore these stats if the call fails)
	if !n.IgnoreProtocolStats {
		netprotos, _ := n.ps.NetProto()
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
	}

	return nil
}

func init() {
	inputs.Add("net", func() telegraf.Input {
		return &NetIOStats{ps: system.NewSystemPS()}
	})
}
