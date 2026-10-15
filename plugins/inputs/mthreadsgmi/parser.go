package mthreadsgmi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs/nvidia_smi/common"
)

const measurement = "mthreads_gmi"

const deviceTypePhysical = "Physical"

func parse(acc telegraf.Accumulator, buf []byte) error {
	gpus, err := decodeGPUs(buf)
	if err != nil {
		return err
	}
	for _, gpu := range gpus {
		if !strings.EqualFold(jsonString(gpu["DeviceType"]), deviceTypePhysical) {
			continue
		}
		tags, fields := gpuMetric(gpu)
		if len(fields) == 0 {
			continue
		}
		acc.AddFields(measurement, fields, tags)
	}
	return nil
}

func decodeGPUs(buf []byte) ([]map[string]interface{}, error) {
	var root struct {
		GPU json.RawMessage `json:"GPU"`
	}
	if err := json.Unmarshal(buf, &root); err != nil {
		return nil, err
	}
	raw := bytes.TrimSpace(root.GPU)
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	switch raw[0] {
	case '[':
		var gpus []map[string]interface{}
		if err := json.Unmarshal(raw, &gpus); err != nil {
			return nil, err
		}
		return gpus, nil
	case '{':
		var gpu map[string]interface{}
		if err := json.Unmarshal(raw, &gpu); err != nil {
			return nil, err
		}
		return []map[string]interface{}{gpu}, nil
	default:
		return nil, fmt.Errorf("unexpected GPU json")
	}
}

func gpuMetric(gpu map[string]interface{}) (map[string]string, map[string]interface{}) {
	tags := map[string]string{}
	common.SetTagIfUsed(tags, "index", jsonString(gpu["Index"]))
	common.SetTagIfUsed(tags, "name", jsonString(gpu["Product Name"]))
	common.SetTagIfUsed(tags, "uuid", jsonString(gpu["GPU UUID"]))
	common.SetTagIfUsed(tags, "pstate", jsonString(gpu["Performance State"]))

	pci := asMap(gpu["PCI"])
	common.SetTagIfUsed(tags, "pci_bus_id", jsonString(pci["Bus ID"]))
	link := asMap(asMap(pci["GPU Link Info"]))

	fields := map[string]interface{}{}
	util := asMap(gpu["Utilization"])
	putInt(fields, "utilization_gpu", jsonString(util["Gpu"]))
	putInt(fields, "utilization_memory", jsonString(util["Memory"]))
	putInt(fields, "utilization_encoder", jsonString(util["Encoder"]))
	putInt(fields, "utilization_decoder", jsonString(util["Decoder"]))

	temp := asMap(gpu["Temperature"])
	putInt(fields, "temperature_gpu", jsonString(temp["GPU Current Temp"]))

	fb := asMap(gpu["FB Memory Usage"])
	putInt(fields, "memory_total", jsonString(fb["Total"]))
	putInt(fields, "memory_used", jsonString(fb["Used"]))
	putInt(fields, "memory_free", jsonString(fb["Free"]))

	power := asMap(gpu["Power Readings"])
	putFloat(fields, "power_draw", mapString(power, "Power Draw"))
	putFloat(fields, "power_limit", jsonString(power["Current Power Limit"]))

	clocks := asMap(gpu["Clocks"])
	putInt(fields, "clocks_current_graphics", jsonString(clocks["Graphics"]))
	putInt(fields, "clocks_current_memory", jsonString(clocks["Memory"]))

	gen := asMap(link["PCIe Generation"])
	width := asMap(link["Link Width"])
	putInt(fields, "pcie_link_gen_current", jsonString(gen["Current"]))
	putInt(fields, "pcie_link_width_current", jsonString(width["Current"]))
	return tags, fields
}

func asMap(v interface{}) map[string]interface{} {
	m, _ := v.(map[string]interface{})
	return m
}

func mapString(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key]; ok {
		if s := jsonString(v); s != "" {
			return s
		}
	}
	for mk, v := range m {
		if strings.TrimSpace(mk) == key {
			return jsonString(v)
		}
	}
	return ""
}

func jsonString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case nil:
		return ""
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}

func parseMetricNumber(raw string) (float64, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.EqualFold(raw, "N/A") {
		return 0, false
	}
	raw = strings.TrimSuffix(raw, "%")
	raw = strings.TrimSuffix(raw, "MiB")
	raw = strings.TrimSuffix(raw, "MHz")
	raw = strings.TrimSuffix(raw, "W")
	raw = strings.TrimSuffix(raw, "C")
	raw = strings.TrimSuffix(raw, "x")
	raw = strings.TrimSpace(raw)
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

func putInt(fields map[string]interface{}, key, raw string) {
	f, ok := parseMetricNumber(raw)
	if !ok {
		return
	}
	fields[key] = int(f)
}

func putFloat(fields map[string]interface{}, key, raw string) {
	f, ok := parseMetricNumber(raw)
	if !ok {
		return
	}
	fields[key] = f
}
