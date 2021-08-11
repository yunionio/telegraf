package swap

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/mem"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/system"
)

type SwapStats struct {
	ps system.PS

	lastStats *mem.SwapMemoryStat
	lastTime  time.Time
}

func (ss *SwapStats) Description() string {
	return "Read metrics about swap memory usage"
}

func (ss *SwapStats) SampleConfig() string { return "" }

func (ss *SwapStats) Gather(acc telegraf.Accumulator) error {
	swap, err := ss.ps.SwapStat()
	if err != nil {
		return fmt.Errorf("error getting swap memory info: %s", err)
	}

	curr := time.Now()
	timeDelta := curr.Sub(ss.lastTime).Seconds()

	fieldsG := map[string]interface{}{
		"total":        swap.Total,
		"used":         swap.Used,
		"free":         swap.Free,
		"used_percent": swap.UsedPercent,
	}
	fieldsC := map[string]interface{}{
		"in":  swap.Sin,
		"out": swap.Sout,
	}
	acc.AddGauge("swap", fieldsG, nil)
	acc.AddCounter("swap", fieldsC, nil)

	if ss.lastStats != nil {
		fields2 := map[string]interface{}{
			"in_bps":  float64(swap.Sin-ss.lastStats.Sin) / timeDelta,
			"out_bps": float64(swap.Sout-ss.lastStats.Sout) / timeDelta,
		}
		acc.AddGauge("swap", fields2, nil, curr)
	}

	ss.lastStats = swap
	ss.lastTime = curr

	return nil
}

func init() {
	ps := system.NewSystemPS()
	inputs.Add("swap", func() telegraf.Input {
		return &SwapStats{ps: ps}
	})
}
