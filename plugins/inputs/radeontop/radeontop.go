//go:generate ../../../tools/readme_config_includer/generator
package radeontop

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
	inputs.Add("radeontop", func() telegraf.Input {
		return &Radeontop{
			BinPath:     "/usr/bin/radeontop",
			DevicePaths: []string{},
			Timeout:     config.Duration(5 * time.Second),
		}
	})
}

const measurement = "radeontop"

//go:embed sample.conf
var sampleConfig string

type Radeontop struct {
	BinPath     string
	DevicePaths []string
	Timeout     config.Duration
}

func (rtp *Radeontop) Description() string {
	return "Pull statistics from AMD GPUS attached to the host"
}

func (rtp *Radeontop) SampleConfig() string {
	return sampleConfig
}

func (rtp *Radeontop) Gather(acc telegraf.Accumulator) error {
	if _, err := procutils.IsRemoteFileExist(rtp.BinPath); err != nil {
		return err
	}

	if len(rtp.DevicePaths) == 0 {
		return errors.Errorf("device_paths are empty")
	}

	results := make([]*result, 0)
	for _, dp := range rtp.DevicePaths {
		result, err := rtp.getDeviceResult(dp)
		if err != nil {
			return fmt.Errorf("get device %s result: %v", dp, err)
		}
		results = append(results, result)
	}

	if err := collectResult(acc, results); err != nil {
		return errors.Wrap(err, "collect result")
	}
	return nil
}

func collectResult(acc telegraf.Accumulator, results []*result) error {
	for _, result := range results {
		acc.AddFields(measurement, result.getFields(), result.getTags())
	}
	return nil
}

func (rtp *Radeontop) getDeviceResult(devPath string) (*result, error) {
	data, err := rtp.pollData(devPath)
	if err != nil {
		return nil, fmt.Errorf("poll data: %v", err)
	}
	return gatherRadeontop(data, devPath)
}

func (rtp *Radeontop) pollData(devPath string) ([]byte, error) {
	//ret, err := internal.CombinedOutputTimeout(
	//	exec.Command(rtp.BinPath, "-p", devPath, "-l", "1", "-d", "-"),
	//	time.Duration(rtp.Timeout))
	ret, err := procutils.NewRemoteCommandAsFarAsPossible(rtp.BinPath, "-p", devPath, "-l", "1", "-d", "-").Output()
	if err != nil {
		return nil, errors.Wrapf(err, "radeontop polling %s failed, out: %s", devPath, ret)
	}
	return ret, nil
}

func gatherRadeontop(data []byte, devPath string) (*result, error) {
	var line string
	for _, l := range strings.Split(string(data), "\n") {
		if strings.Contains(l, "Dumping to") {
			continue
		}
		if len(l) == 0 {
			continue
		}
		line = l
	}
	result, err := parseRadeontopLine(line)
	if err != nil {
		return nil, fmt.Errorf("radeontop parsing failed: %v", err)
	}
	result.DevicePath = devPath
	return result, nil
}

type result struct {
	DevicePath string
	Bus        string
	// Graphics pipe
	GPU float64
	// Event engine: ee
	EventEngine float64
	// Vertex Grouper + Tesselator: vgt
	VertexGrouperTesselator float64
	// ta
	TextureAddresser float64
	// sx
	ShaderExport float64
	// sh
	SequencerInstructionCache float64
	// spi
	ShaderInterpolator float64
	// sc
	ScanConverter float64
	// pa
	PrimitiveAssembly float64
	// db
	DepthBlock float64
	// cb
	ColorBlock float64
	// vram
	VRAM           float64
	VRAMUsedSizeMB int64
	VRAMSizeMB     int64
	// gtt
	GTT           float64
	GTTUsedSizeMB int64
	GTTSizeMB     int64
	// mclk ghz
	MemoryClock    float64
	MemoryClockMHz float64
	// sclk ghz
	ShaderClock    float64
	ShaderClockMHz float64
}

func (r result) getFields() map[string]interface{} {
	return map[string]interface{}{
		"memory_total":                            r.VRAMSizeMB,
		"memory_used":                             r.VRAMUsedSizeMB,
		"memory_free":                             r.VRAMSizeMB - r.VRAMUsedSizeMB,
		"gtt_total":                               r.GTTSizeMB,
		"gtt_free":                                r.GTTSizeMB - r.GTTUsedSizeMB,
		"gtt_used":                                r.GTTUsedSizeMB,
		"utilization_gpu":                         r.GPU,
		"utilization_memory":                      r.VRAM,
		"utilization_gtt":                         r.GTT,
		"utilization_event_engine":                r.EventEngine,
		"utilization_vertex_grouper_tesselator":   r.VertexGrouperTesselator,
		"utilization_texture_addresser":           r.TextureAddresser,
		"utilization_shader_export":               r.ShaderExport,
		"utilization_sequencer_instruction_cache": r.SequencerInstructionCache,
		"utilization_shader_interpolator":         r.ShaderInterpolator,
		"utilization_scan_converter":              r.ScanConverter,
		"utilization_primitive_assembly":          r.PrimitiveAssembly,
		"utilization_depth_block":                 r.DepthBlock,
		"utilization_color_block":                 r.ColorBlock,
		"clocks_current_memory":                   r.MemoryClockMHz,
		"utilization_clock_memory":                r.MemoryClock,
		"utilization_clock_shader":                r.ShaderClock,
		"clocks_current_shader":                   r.ShaderClockMHz,
	}
}

func (r result) getTags() map[string]string {
	return map[string]string{
		"device_path": r.DevicePath,
		"bus":         r.Bus,
	}
}

type fillFunc func(part string, r *result) error

func getString2Value(part string) (string, error) {
	data := strings.Split(part, " ")
	if len(data) != 2 {
		return "", fmt.Errorf("wrong number of fields: %s", part)
	}
	return data[1], nil
}

func roundFloat(f float64) float64 {
	return math.Round(f*100) / 100
}

func getFloatValue(value string, trimUnit string) (float64, error) {
	value = strings.TrimRight(value, trimUnit)
	result, err := strconv.ParseFloat(value, 32)
	if err != nil {
		return 0, fmt.Errorf("wrong value for %s, unit %s: %v", value, trimUnit, err)
	}
	return roundFloat(result), nil
}

func getPercent2Value(part string) (float64, error) {
	data := strings.Split(part, " ")
	if len(data) != 2 {
		return 0, fmt.Errorf("wrong number of fields: %s", part)
	}
	result, err := getFloatValue(data[1], "%")
	if err != nil {
		return 0, fmt.Errorf("wrong part %s: %v", part, err)
	}
	return result, nil
}

func getPercent2ValueWithResult(part string, f func(percent float64)) error {
	percent, err := getPercent2Value(part)
	if err != nil {
		return err
	}
	f(percent)
	return nil
}

func getPercent3Value(part string, ignoreUnit1, ignoreUnit2 string) (float64, float64, error) {
	data := strings.Split(part, " ")
	if len(data) != 3 {
		return 0, 0, fmt.Errorf("wrong number of fields: %s", part)
	}
	result1, err := getFloatValue(data[1], ignoreUnit1)
	if err != nil {
		return 0, 0, fmt.Errorf("wrong part %s: %v", part, err)
	}
	result2, err := getFloatValue(data[2], ignoreUnit2)
	if err != nil {
		return 0, 0, fmt.Errorf("wrong part %s: %v", part, err)
	}
	return result1, result2, nil
}

func getPercent3ValueWithResult(part string, unit2 string, f func(percent, extraValue float64)) error {
	percent, extraValue, err := getPercent3Value(part, "%", unit2)
	if err != nil {
		return err
	}
	f(percent, extraValue)
	return nil
}

var fillFuncs = map[string]fillFunc{
	"bus": func(part string, r *result) error {
		bus, err := getString2Value(part)
		if err != nil {
			return err
		}
		r.Bus = bus
		return nil
	},
	"gpu": func(part string, r *result) error {
		return getPercent2ValueWithResult(part, func(percent float64) {
			r.GPU = percent
		})
	},
	"ee": func(part string, r *result) error {
		return getPercent2ValueWithResult(part, func(percent float64) {
			r.EventEngine = percent
		})
	},
	"vgt": func(part string, r *result) error {
		return getPercent2ValueWithResult(part, func(percent float64) {
			r.VertexGrouperTesselator = percent
		})
	},
	"ta": func(part string, r *result) error {
		return getPercent2ValueWithResult(part, func(percent float64) {
			r.TextureAddresser = percent
		})
	},
	"sx": func(part string, r *result) error {
		return getPercent2ValueWithResult(part, func(percent float64) {
			r.ShaderExport = percent
		})
	},
	"sh": func(part string, r *result) error {
		return getPercent2ValueWithResult(part, func(percent float64) {
			r.SequencerInstructionCache = percent
		})
	},
	"spi": func(part string, r *result) error {
		return getPercent2ValueWithResult(part, func(percent float64) {
			r.ShaderInterpolator = percent
		})
	},
	"sc": func(part string, r *result) error {
		return getPercent2ValueWithResult(part, func(percent float64) {
			r.ScanConverter = percent
		})
	},
	"pa": func(part string, r *result) error {
		return getPercent2ValueWithResult(part, func(percent float64) {
			r.PrimitiveAssembly = percent
		})
	},
	"db": func(part string, r *result) error {
		return getPercent2ValueWithResult(part, func(percent float64) {
			r.DepthBlock = percent
		})
	},
	"cb": func(part string, r *result) error {
		return getPercent2ValueWithResult(part, func(percent float64) {
			r.ColorBlock = percent
		})
	},
	"vram": func(part string, r *result) error {
		return getPercent3ValueWithResult(part, "mb", func(percent, usedMb float64) {
			r.VRAM = percent
			r.VRAMUsedSizeMB = int64(usedMb)
			r.VRAMSizeMB = int64(usedMb / (percent / 100))
		})
	},
	"gtt": func(part string, r *result) error {
		return getPercent3ValueWithResult(part, "mb", func(percent, usedMb float64) {
			r.GTT = percent
			r.GTTUsedSizeMB = int64(usedMb)
			r.GTTSizeMB = int64(usedMb / (percent / 100))
		})
	},
	"mclk": func(part string, r *result) error {
		return getPercent3ValueWithResult(part, "ghz", func(percent, gzh float64) {
			r.MemoryClock = percent
			r.MemoryClockMHz = gzh * 1000
		})
	},
	"sclk": func(part string, r *result) error {
		return getPercent3ValueWithResult(part, "ghz", func(percent, gzh float64) {
			r.ShaderClock = percent
			r.ShaderClockMHz = gzh * 1000
		})
	},
}

// 1731392941.164959: bus 03, gpu 14.17%, ee 0.00%, vgt 7.50%, ta 4.17%, sx 4.17%, sh 7.50%, spi 8.33%, sc 5.83%, pa 2.50%, db 5.83%, cb 5.83%, vram 62.58% 20482.97mb, gtt 15.94% 5218.41mb, mclk 100.00% 1.000ghz, sclk 94.13% 2.405ghz
func parseRadeontopLine(line string) (*result, error) {
	parts := strings.Split(line, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid radeontop line: %s", line)
	}
	// drop timestamp
	metricLine := parts[1]
	metricParts := strings.Split(metricLine, ",")
	result := &result{}
	for _, metricPart := range metricParts {
		metricPart = strings.TrimSpace(metricPart)
		for prefix, f := range fillFuncs {
			if strings.HasPrefix(metricPart, prefix+" ") {
				if err := f(metricPart, result); err != nil {
					return nil, fmt.Errorf("parse metric part %s with %s", prefix, metricPart)
				}
				break
			}
		}
	}
	return result, nil
}
