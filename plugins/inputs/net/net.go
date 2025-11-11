//go:generate ../../../tools/readme_config_includer/generator
package net

import (
	_ "embed"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	pnet "github.com/shirou/gopsutil/v4/net"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/psutil"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type InterfaceProfile struct {
	Name  string
	Alias string
	Speed int
}

type Net struct {
	Interfaces          []string `toml:"interfaces"`
	IgnoreProtocolStats bool     `toml:"ignore_protocol_stats"`

	InterfaceConf []InterfaceProfile

	lastTime  time.Time
	lastStats map[string]pnet.IOCountersStat

	filter     filter.Filter
	ps         psutil.PS
	skipChecks bool
}

func (*Net) SampleConfig() string {
	return sampleConfig
}

func (s *Net) getProfile(name string) *InterfaceProfile {
	for _, inf := range s.InterfaceConf {
		if inf.Name == name {
			return &inf
		}
	}
	return nil
}

func (n *Net) Init() error {
	if !n.IgnoreProtocolStats {
		config.PrintOptionValueDeprecationNotice("inputs.net", "ignore_protocol_stats", "false",
			telegraf.DeprecationInfo{
				Since:     "1.27.3",
				RemovalIn: "1.36.0",
				Notice:    "use the 'inputs.nstat' plugin instead for protocol stats",
			},
		)
	}

	// So not use the interface list of the system if the HOST_PROC variable is
	// set as the interfaces are determined by a syscall and therefore might
	// differ especially in container environments.
	n.skipChecks = os.Getenv("HOST_PROC") != ""

	n.lastStats = make(map[string]pnet.IOCountersStat)
	n.lastTime = time.Now()
	return nil
}

func (n *Net) Gather(acc telegraf.Accumulator) error {
	netio, err := n.ps.NetIO()
	if err != nil {
		return fmt.Errorf("error getting net io info: %w", err)
	}

	if n.filter == nil {
		if n.filter, err = filter.Compile(n.Interfaces); err != nil {
			return fmt.Errorf("error compiling filter: %w", err)
		}
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return fmt.Errorf("error getting list of interfaces: %w", err)
	}
	interfacesByName := make(map[string]net.Interface, len(interfaces))
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
			"speed":        getInterfaceSpeed(io.Name),
		}

		if last, ok := n.lastStats[io.Name]; ok {
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
				speed = int(getInterfaceSpeed(io.Name))
			}
			fields2["speed"] = speed
			if speed > 0 {
				fields2["if_in_percent"] = bpsRecv / float64(speed) / 10000.0
				fields2["if_out_percent"] = bpsSent / float64(speed) / 10000.0
			}

			for k, v := range fields2 {
				fields[k] = v
			}
		}
		n.lastStats[io.Name] = io
		acc.AddCounter("net", fields, tags)
	}
	n.lastTime = curr

	// Get system wide stats for different network protocols
	// (ignore these stats if the call fails)
	if !n.IgnoreProtocolStats {
		//nolint:errcheck // stats ignored on fail
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

// Get the interface speed from /sys/class/net/*/speed file. returns -1 if unsupported
func getInterfaceSpeed(ioName string) int64 {
	sysPath := internal.GetSysPath()

	raw, err := os.ReadFile(filepath.Join(sysPath, "class", "net", ioName, "speed"))
	if err != nil {
		return -1
	}

	speed, err := strconv.ParseInt(strings.TrimSuffix(string(raw), "\n"), 10, 64)
	if err != nil {
		return -1
	}
	return speed
}

func init() {
	inputs.Add("net", func() telegraf.Input {
		return &Net{ps: psutil.NewSystemPS()}
	})
}
