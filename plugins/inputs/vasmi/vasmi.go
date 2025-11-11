//go:generate ../../../tools/readme_config_includer/generator
package vasmi

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
	inputs.Add("vasmi", func() telegraf.Input {
		return &vasmi{
			BinPath: "/usr/bin/vasmi",
			Timeout: config.Duration(5 * time.Second),
		}
	})
}

const meaurement = "vasmi"

//go:embed sample.conf
var sampleConfig string

type vasmi struct {
	BinPath string
	Timeout config.Duration
}

func (v vasmi) SampleConfig() string {
	return sampleConfig
}

func (v vasmi) Description() string {
	return "Pull statistics from vasmi dmon attached to the host"
}

func (v vasmi) Gather(acc telegraf.Accumulator) error {
	if _, err := procutils.IsRemoteFileExist(v.BinPath); err != nil {
		return err
	}
	results, err := v.pollMetrics()
	if err != nil {
		return errors.Wrap(err, "pollMetrics")
	}
	if err := colloectResults(acc, results); err != nil {
		return errors.Wrap(err, "collect result")
	}
	return nil
}

func colloectResults(acc telegraf.Accumulator, aics []*AIC) error {
	for _, aic := range aics {
		for _, dev := range aic.Devices {
			acc.AddFields(meaurement, dev.getFields(aic), dev.getTags(aic))
		}
	}
	return nil
}

type AIC struct {
	AIC     string    `json:"aic"`
	Power   float64   `json:"power"`
	Devices []*Device `json:"devices"`
}

var (
	AicLineLen = 17
	DevLineLen = 14
)

func roundFloat(f float64) float64 {
	return math.Round(f*100) / 100
}

func getFloatValue(value string) (float64, error) {
	if value == "N/A" {
		return 0.0, nil
	}
	result, err := strconv.ParseFloat(value, 32)
	if err != nil {
		return 0, fmt.Errorf("wrong value for %s: %v", value, err)
	}
	return roundFloat(result), nil
}

func newAIC(aicParts []string) (*AIC, error) {
	if len(aicParts) != AicLineLen {
		return nil, errors.Errorf("wrong aic parts: %v", aicParts)
	}
	aic := &AIC{
		AIC:     aicParts[0],
		Devices: make([]*Device, 0),
	}
	powerStr := aicParts[1]
	powerVal, err := getFloatValue(powerStr)
	if err != nil {
		return nil, errors.Wrapf(err, "get power with: %s", powerStr)
	}
	aic.Power = powerVal
	devParts := aicParts[3:]
	dev, err := newDevice(devParts)
	if err != nil {
		return nil, errors.Wrapf(err, "new device with: %v", devParts)
	}
	aic.Devices = append(aic.Devices, dev)
	return aic, nil
}

type Device struct {
	DevId    string  `json:"dev_id"`
	DieId    string  `json:"die_id"`
	Temp     float64 `json:"temp"`
	Oclk     float64 `json:"oclk"` // MHz
	Dclk     float64 `json:"dclk"` // MHz
	Eclk     float64 `json:"eclk"`
	Gclk     float64 `json:"gclk"`
	ShareMem float64 `json:"share_mem"`
	GPU      float64 `json:"gpu"`
	Dec      float64 `json:"dec"`
	Enc      float64 `json:"enc"`
	AI       float64 `json:"ai"`
	GfxMem   float64 `json:"gfx_mem"`
}

func newDevice(devParts []string) (*Device, error) {
	if len(devParts) != DevLineLen {
		return nil, errors.Errorf("wrong device parts: %v", devParts)
	}
	dev := &Device{
		DevId: devParts[0],
		DieId: devParts[1],
	}

	errs := []error{}
	iF := func(part string, f func(val float64)) {
		val, err := getFloatValue(part)
		if err != nil {
			errs = append(errs, errors.Wrapf(err, "get float value: %s", part))
			return
		}
		f(val)
	}

	iF(devParts[2], func(val float64) {
		dev.Temp = val
	})
	iF(devParts[3], func(val float64) {
		dev.Oclk = val
	})
	iF(devParts[4], func(val float64) {
		dev.Dclk = val
	})
	iF(devParts[5], func(val float64) {
		dev.Eclk = val
	})
	iF(devParts[6], func(val float64) {
		dev.Gclk = val
	})
	iF(devParts[7], func(val float64) {
		dev.ShareMem = val
	})
	iF(devParts[8], func(val float64) {
		dev.GPU = val
	})
	iF(devParts[9], func(val float64) {
		dev.Dec = val
	})
	iF(devParts[10], func(val float64) {
		dev.Enc = val
	})
	iF(devParts[11], func(val float64) {
		dev.AI = val
	})
	iF(devParts[12], func(val float64) {
		dev.GfxMem = val
	})

	if len(errs) > 0 {
		var msg string
		for _, err := range errs {
			msg += err.Error() + "\n"
		}
		return nil, errors.New(msg)
	}
	return dev, nil
}

func (v vasmi) pollMetrics() ([]*AIC, error) {
	//ret, err := internal.CombinedOutputTimeout(
	//	exec.Command(v.BinPath, "dmon", "--loop", "1"),
	//	time.Duration(v.Timeout))
	ret, err := procutils.NewRemoteCommandAsFarAsPossible(v.BinPath, "dmon", "--loop", "1").Output()
	if err != nil {
		return nil, errors.Wrapf(err, "%s dmon --loop 1: %s", v.BinPath, ret)
	}
	return parseResults(ret)
}

func parseResults(content []byte) ([]*AIC, error) {
	lines := strings.Split(string(content), "\n")
	startIdx := 0
	endIdx := 0
	for i, line := range lines {
		if strings.HasPrefix(line, "---------") {
			startIdx = i + 1
		}
		if strings.HasPrefix(line, "<1") {
			endIdx = i
			break
		}
	}
	if startIdx >= endIdx {
		return nil, errors.Errorf("can't found start(%d) end end(%d) index", startIdx, endIdx)
	}
	filterLines := lines[startIdx:endIdx]
	aics := make([]*AIC, 0)
	var tmpAic *AIC
	for _, line := range filterLines {
		parts := strings.Fields(line)
		if len(parts) == AicLineLen {
			aic, err := newAIC(parts)
			if err != nil {
				return nil, errors.Wrapf(err, "new aic with: %v", parts)
			}
			aics = append(aics, aic)
			tmpAic = aic
		} else if len(parts) == DevLineLen {
			dev, err := newDevice(parts)
			if err != nil {
				return nil, errors.Wrapf(err, "new device with: %v", parts)
			}
			if tmpAic == nil {
				return nil, errors.Errorf("aic is nil")
			}
			tmpAic.Devices = append(tmpAic.Devices, dev)
		} else {
			return nil, errors.Errorf("invalid line: %v", line)
		}
	}
	return aics, nil
}

func (d Device) getFields(aic *AIC) map[string]interface{} {
	return map[string]interface{}{
		"aic_power":                aic.Power,
		"temperature_gpu":          d.Temp,
		"utilization_gpu":          d.GPU,
		"utilization_memory":       d.GfxMem,
		"utilization_share_memory": d.ShareMem,
		"utilization_encoder":      d.Enc,
		"utilization_decoder":      d.Dec,
		"utilization_ai":           d.AI,
		"clocks_current_gpu":       d.Gclk,
		"oclk":                     d.Oclk,
		"dclk":                     d.Dclk,
		"eclk":                     d.Eclk,
		"gclk":                     d.Gclk,
	}
}

func (d Device) getTags(aic *AIC) map[string]string {
	return map[string]string{
		"aic":    aic.AIC,
		"dev_id": d.DevId,
		"die_id": d.DieId,
	}
}
