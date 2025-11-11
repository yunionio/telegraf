//go:generate ../../../tools/readme_config_includer/generator
package swap

import (
	_ "embed"
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/psutil"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/shirou/gopsutil/v4/mem"
)

//go:embed sample.conf
var sampleConfig string

type Swap struct {
	ps psutil.PS

	lastStats *mem.SwapMemoryStat
	lastTime  time.Time
}

func (*Swap) SampleConfig() string {
	return sampleConfig
}

func (ss *Swap) Gather(acc telegraf.Accumulator) error {
	swap, err := ss.ps.SwapStat()
	if err != nil {
		return fmt.Errorf("error getting swap memory info: %w", err)
	}
	curr := time.Now()
	timeDelta := curr.Sub(ss.lastTime).Seconds()

	fieldsG := map[string]interface{}{
		"total":        swap.Total,
		"used":         swap.Used,
		"free":         swap.Free,
		"used_percent": swap.UsedPercent,
	}

	if ss.lastStats != nil {
		fieldsG["in_bps"] = float64(swap.Sin-ss.lastStats.Sin) / timeDelta
		fieldsG["out_bps"] = float64(swap.Sout-ss.lastStats.Sout) / timeDelta
	}
	ss.lastStats = swap
	ss.lastTime = curr

	fieldsC := map[string]interface{}{
		"in":  swap.Sin,
		"out": swap.Sout,
	}
	acc.AddGauge("swap", fieldsG, nil)
	acc.AddCounter("swap", fieldsC, nil)

	return nil
}

func init() {
	ps := psutil.NewSystemPS()
	inputs.Add("swap", func() telegraf.Input {
		return &Swap{ps: ps}
	})
}
