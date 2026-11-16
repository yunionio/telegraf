//go:generate ../../../tools/readme_config_includer/generator
package brsmi

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/procutils"
	"github.com/influxdata/telegraf/plugins/inputs"
)

func init() {
	inputs.Add("brsmi", func() telegraf.Input {
		return &brsmi{BinPath: "/usr/bin/brsmi", Timeout: config.Duration(5 * time.Second)}
	})
}

const measurement = "brsmi"

//go:embed sample.conf
var sampleConfig string

type brsmi struct {
	BinPath string
	Timeout config.Duration
}

func (b brsmi) SampleConfig() string { return sampleConfig }

func (b brsmi) Description() string { return "Pull GPU statistics from brsmi" }

func (b brsmi) Gather(acc telegraf.Accumulator) error {
	if _, err := procutils.IsRemoteFileExist(b.BinPath); err != nil {
		return err
	}
	metrics, err := b.pollMetrics()
	if err != nil {
		return errors.Wrap(err, "pollMetrics")
	}
	for _, metric := range metrics {
		acc.AddFields(measurement, metric.fields(), metric.tags())
	}
	return nil
}

type GPU struct {
	Index int
	Power float64
	GTemp float64
	MTemp float64
	SPC   float64
	Mem   float64
	Enc   float64
	Dec   float64
	MClk  float64
	PClk  float64
}

func (g GPU) fields() map[string]interface{} {
	return map[string]interface{}{
		"power": g.Power, "temperature_gpu": g.GTemp, "temperature_memory": g.MTemp,
		"utilization_spc": g.SPC, "utilization_memory": g.Mem,
		"utilization_encoder": g.Enc, "utilization_decoder": g.Dec,
		"clocks_memory": g.MClk, "clocks_gpu": g.PClk,
	}
}

func (g GPU) tags() map[string]string { return map[string]string{"index": strconv.Itoa(g.Index)} }

func parseResults(content []byte) ([]*GPU, error) {
	metrics := make([]*GPU, 0)
	for _, line := range strings.Split(string(content), "\n") {
		parts := strings.Fields(line)
		if len(parts) == 0 || strings.HasPrefix(parts[0], "#") {
			continue
		}
		if len(parts) != 10 {
			return nil, errors.Errorf("invalid brsmi data row: %q", line)
		}
		index, err := strconv.Atoi(parts[0])
		if err != nil { // Ignore non-data output, but reject malformed numeric rows below.
			continue
		}
		values := make([]float64, 9)
		for i, value := range parts[1:] {
			if strings.Contains(value, "~") { // Range-valued limits (for example 0~85) are not measurements.
				continue
			}
			values[i], err = strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid value %q in brsmi row %q: %w", value, line, err)
			}
		}
		metrics = append(metrics, &GPU{Index: index, Power: values[0], GTemp: values[1], MTemp: values[2], SPC: values[3], Mem: values[4], Enc: values[5], Dec: values[6], MClk: values[7], PClk: values[8]})
	}
	if len(metrics) == 0 {
		return nil, errors.New("no brsmi GPU metrics found")
	}
	return metrics, nil
}

func (b brsmi) pollMetrics() ([]*GPU, error) {
	ret, err := procutils.NewRemoteCommandAsFarAsPossible(b.BinPath, "gpu", "dmon", "-c", "1").Output()
	if err != nil {
		return nil, errors.Wrapf(err, "%s gpu dmon -c 1: %s", b.BinPath, ret)
	}
	return parseResults(ret)
}
