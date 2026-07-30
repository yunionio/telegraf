//go:generate ../../../tools/readme_config_includer/generator
package hysmi

import (
	_ "embed"
	"fmt"
	"math"
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
	inputs.Add("hysmi", func() telegraf.Input {
		return &hysmi{
			BinPath: "/opt/hyhal/bin/hy-smi",
			Timeout: config.Duration(5 * time.Second),
		}
	})
}

const measurement = "hysmi"

const dataLineLen = 10

//go:embed sample.conf
var sampleConfig string

type hysmi struct {
	BinPath string
	Timeout config.Duration
}

func (h hysmi) SampleConfig() string {
	return sampleConfig
}

func (h hysmi) Description() string {
	return "Pull statistics from hy-smi attached to the host"
}

func (h hysmi) Gather(acc telegraf.Accumulator) error {
	if _, err := procutils.IsRemoteFileExist(h.BinPath); err != nil {
		return err
	}
	results, err := h.pollMetrics()
	if err != nil {
		return errors.Wrap(err, "pollMetrics")
	}
	if err := collectResults(acc, results); err != nil {
		return errors.Wrap(err, "collect result")
	}
	return nil
}

func collectResults(acc telegraf.Accumulator, hcus []*HCU) error {
	for _, hcu := range hcus {
		acc.AddFields(measurement, hcu.getFields(), hcu.getTags())
	}
	return nil
}

type HCU struct {
	Index    string  `json:"hcu"`
	Temp     float64 `json:"temp"`
	AvgPwr   float64 `json:"avg_pwr"`
	Perf     string  `json:"perf"`
	PwrCap   float64 `json:"pwr_cap"`
	VRAM     float64 `json:"vram"`
	HCUUtil  float64 `json:"hcu_util"`
	Dec      float64 `json:"dec"`
	Enc      float64 `json:"enc"`
	Mode     string  `json:"mode"`
}

func roundFloat(f float64) float64 {
	return math.Round(f*100) / 100
}

func parseFloatValue(value string) (float64, error) {
	value = strings.TrimSpace(value)
	if value == "" || value == "N/A" {
		return 0.0, nil
	}
	value = strings.TrimSuffix(value, "C")
	value = strings.TrimSuffix(value, "W")
	value = strings.TrimSuffix(value, "%")
	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("wrong value for %s: %v", value, err)
	}
	return roundFloat(result), nil
}

func newHCU(parts []string) (*HCU, error) {
	if len(parts) != dataLineLen {
		return nil, errors.Errorf("wrong hcu parts: %v", parts)
	}
	hcu := &HCU{
		Index: parts[0],
		Perf:  parts[3],
		Mode:  parts[9],
	}
	errs := []error{}
	iF := func(part string, f func(val float64)) {
		val, err := parseFloatValue(part)
		if err != nil {
			errs = append(errs, errors.Wrapf(err, "get float value: %s", part))
			return
		}
		f(val)
	}
	iF(parts[1], func(val float64) { hcu.Temp = val })
	iF(parts[2], func(val float64) { hcu.AvgPwr = val })
	iF(parts[4], func(val float64) { hcu.PwrCap = val })
	iF(parts[5], func(val float64) { hcu.VRAM = val })
	iF(parts[6], func(val float64) { hcu.HCUUtil = val })
	iF(parts[7], func(val float64) { hcu.Dec = val })
	iF(parts[8], func(val float64) { hcu.Enc = val })
	if len(errs) > 0 {
		var msg string
		for _, err := range errs {
			msg += err.Error() + "\n"
		}
		return nil, errors.New(msg)
	}
	return hcu, nil
}

func (h hysmi) pollMetrics() ([]*HCU, error) {
	ret, err := procutils.NewRemoteCommandAsFarAsPossible(h.BinPath).Output()
	if err != nil {
		return nil, errors.Wrapf(err, "%s: %s", h.BinPath, ret)
	}
	return parseResults(ret)
}

func parseResults(content []byte) ([]*HCU, error) {
	lines := strings.Split(string(content), "\n")
	headerFound := false
	hcus := make([]*HCU, 0)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "====") {
			if headerFound {
				break
			}
			continue
		}
		parts := strings.Fields(line)
		if !headerFound {
			if len(parts) >= 2 && parts[0] == "HCU" && parts[1] == "Temp" {
				headerFound = true
			}
			continue
		}
		if _, err := strconv.Atoi(parts[0]); err != nil {
			continue
		}
		hcu, err := newHCU(parts)
		if err != nil {
			return nil, errors.Wrapf(err, "new hcu with: %v", parts)
		}
		hcus = append(hcus, hcu)
	}
	if !headerFound {
		return nil, errors.New("can't found hcu header line")
	}
	if len(hcus) == 0 {
		return nil, errors.New("no hcu data lines found")
	}
	return hcus, nil
}

func (h HCU) getFields() map[string]interface{} {
	return map[string]interface{}{
		"temperature_gpu":     h.Temp,
		"power_draw":          h.AvgPwr,
		"power_cap":           h.PwrCap,
		"utilization_memory":  h.VRAM,
		"utilization_gpu":     h.HCUUtil,
		"utilization_decoder": h.Dec,
		"utilization_encoder": h.Enc,
	}
}

func (h HCU) getTags() map[string]string {
	return map[string]string{
		"hcu":       h.Index,
		"perf_mode": h.Perf,
		"mode":      h.Mode,
	}
}
