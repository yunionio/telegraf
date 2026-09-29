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
	Index        string  `json:"hcu"`
	Temp         float64 `json:"temp"`
	AvgPwr       float64 `json:"avg_pwr"`
	Perf         string  `json:"perf"`
	PwrCap       float64 `json:"pwr_cap"`
	VRAM         float64 `json:"vram"`
	HCUUtil      float64 `json:"hcu_util"`
	Dec          float64 `json:"dec"`
	Enc          float64 `json:"enc"`
	Mode         string  `json:"mode"`
	GttTotal     int     `json:"gtt_total"`
	GttUsed      int     `json:"gtt_used"`
	GttFree      int     `json:"gtt_free"`
	VisVramTotal int     `json:"vis_vram_total"`
	VisVramUsed  int     `json:"vis_vram_used"`
	VisVramFree  int     `json:"vis_vram_free"`
	VramTotal    int     `json:"vram_total"`
	VramUsed     int     `json:"vram_used"`
	VramFree     int     `json:"vram_free"`
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
	hcus, err := parseResults(ret)
	if err != nil {
		return nil, err
	}
	// Best-effort memory info; ignore errors to keep backward compatibility.
	memRet, err := procutils.NewRemoteCommandAsFarAsPossible(h.BinPath, "--showmeminfo", "all").Output()
	if err == nil {
		_ = parseMemoryInfo(hcus, memRet)
	}
	return hcus, nil
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

func parseMemoryInfo(hcus []*HCU, content []byte) error {
	hcuMap := make(map[string]*HCU, len(hcus))
	for _, hcu := range hcus {
		hcuMap[hcu.Index] = hcu
	}

	for line := range strings.SplitSeq(string(content), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "HCU[") {
			continue
		}
		parts := strings.SplitN(line, ":", 3)
		if len(parts) != 3 {
			continue
		}

		indexPart := strings.TrimSpace(parts[0])
		indexPart = strings.TrimPrefix(indexPart, "HCU[")
		indexPart = strings.TrimSuffix(indexPart, "]")
		hcu, ok := hcuMap[indexPart]
		if !ok {
			continue
		}

		mid := strings.TrimSpace(parts[1])
		mid = strings.TrimSuffix(mid, " (MiB)")
		var memType, field string
		if strings.HasSuffix(mid, "Total Used Memory") {
			field = "used"
			memType = strings.TrimSpace(strings.TrimSuffix(mid, "Total Used Memory"))
		} else if strings.HasSuffix(mid, "Total Memory") {
			field = "total"
			memType = strings.TrimSpace(strings.TrimSuffix(mid, "Total Memory"))
		} else {
			continue
		}

		valueStr := strings.TrimSpace(parts[2])
		value, err := strconv.Atoi(valueStr)
		if err != nil {
			continue
		}

		switch memType {
		case "gtt":
			if field == "total" {
				hcu.GttTotal = value
			} else {
				hcu.GttUsed = value
			}
		case "vis_vram":
			if field == "total" {
				hcu.VisVramTotal = value
			} else {
				hcu.VisVramUsed = value
			}
		case "vram":
			if field == "total" {
				hcu.VramTotal = value
			} else {
				hcu.VramUsed = value
			}
		}
	}

	for _, hcu := range hcus {
		hcu.GttFree = max(0, hcu.GttTotal-hcu.GttUsed)
		hcu.VisVramFree = max(0, hcu.VisVramTotal-hcu.VisVramUsed)
		hcu.VramFree = max(0, hcu.VramTotal-hcu.VramUsed)
	}
	return nil
}

func (h HCU) getFields() map[string]interface{} {
	return map[string]interface{}{
		"temperature_gpu":       h.Temp,
		"power_draw":            h.AvgPwr,
		"power_cap":             h.PwrCap,
		"utilization_memory":    h.VRAM,
		"utilization_gpu":       h.HCUUtil,
		"utilization_decoder":   h.Dec,
		"utilization_encoder":   h.Enc,
		"memory_total":          h.VramTotal,
		"memory_used":           h.VramUsed,
		"memory_free":           h.VramFree,
		"memory_gtt_total":      h.GttTotal,
		"memory_gtt_used":       h.GttUsed,
		"memory_gtt_free":       h.GttFree,
		"memory_vis_vram_total": h.VisVramTotal,
		"memory_vis_vram_used":  h.VisVramUsed,
		"memory_vis_vram_free":  h.VisVramFree,
	}
}

func (h HCU) getTags() map[string]string {
	return map[string]string{
		"hcu":       h.Index,
		"perf_mode": h.Perf,
		"mode":      h.Mode,
	}
}
