// +build linux

package system

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// /proc/stat file line prefixes to gather stats on:
var (
	interrupts       = []byte("intr")
	context_switches = []byte("ctxt")
	processes_forked = []byte("processes")
	disk_pages       = []byte("page")
	boot_time        = []byte("btime")
)

type kernelStats struct {
	ctx int64
	intr int64
	proc int64
	pageIn int64
	pageOut int64
}

type Kernel struct {
	statFile string

	lastTime time.Time
	lastStats *kernelStats
}

func (k *Kernel) Description() string {
	return "Get kernel statistics from /proc/stat"
}

func (k *Kernel) SampleConfig() string { return "" }

func (k *Kernel) Gather(acc telegraf.Accumulator) error {
	data, err := k.getProcStat()
	if err != nil {
		return err
	}

	curr := time.Now()
	timeDelta := curr.Sub(k.lastTime).Seconds()
	stat := kernelStats{}

	fields := make(map[string]interface{})

	dataFields := bytes.Fields(data)
	for i, field := range dataFields {
		switch {
		case bytes.Equal(field, interrupts):
			m, err := strconv.ParseInt(string(dataFields[i+1]), 10, 64)
			if err != nil {
				return err
			}
			fields["interrupts"] = int64(m)
			stat.intr = int64(m)
		case bytes.Equal(field, context_switches):
			m, err := strconv.ParseInt(string(dataFields[i+1]), 10, 64)
			if err != nil {
				return err
			}
			fields["context_switches"] = int64(m)
			stat.ctx = int64(m)
		case bytes.Equal(field, processes_forked):
			m, err := strconv.ParseInt(string(dataFields[i+1]), 10, 64)
			if err != nil {
				return err
			}
			fields["processes_forked"] = int64(m)
			stat.proc = int64(m)
		case bytes.Equal(field, boot_time):
			m, err := strconv.ParseInt(string(dataFields[i+1]), 10, 64)
			if err != nil {
				return err
			}
			fields["boot_time"] = int64(m)
		case bytes.Equal(field, disk_pages):
			in, err := strconv.ParseInt(string(dataFields[i+1]), 10, 64)
			if err != nil {
				return err
			}
			out, err := strconv.ParseInt(string(dataFields[i+2]), 10, 64)
			if err != nil {
				return err
			}
			fields["disk_pages_in"] = int64(in)
			fields["disk_pages_out"] = int64(out)
			stat.pageIn = int64(in)
			stat.pageOut = int64(out)
		}
	}

	acc.AddCounter("kernel", fields, map[string]string{}, curr)

	if k.lastStats != nil {
		fields2 := map[string]interface{} {
			"ctxs_rate": float64(stat.ctx - k.lastStats.ctx)/timeDelta,
			"intr_rate": float64(stat.intr - k.lastStats.intr)/timeDelta,
			"proc_rate": float64(stat.proc - k.lastStats.proc)/timeDelta,
			"pagein_rate": float64(stat.pageIn - k.lastStats.pageIn)/timeDelta,
			"pageout_rate": float64(stat.pageOut - k.lastStats.pageOut)/timeDelta,
		}
		acc.AddGauge("kernel", fields2, map[string]string{}, curr)
	}

	k.lastTime = curr
	k.lastStats = &stat

	return nil
}

func (k *Kernel) getProcStat() ([]byte, error) {
	if _, err := os.Stat(k.statFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("kernel: %s does not exist!", k.statFile)
	} else if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(k.statFile)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func init() {
	inputs.Add("kernel", func() telegraf.Input {
		return &Kernel{
			statFile: "/proc/stat",
		}
	})
}
