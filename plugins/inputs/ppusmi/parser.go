package ppusmi

import (
	"bytes"
	"encoding/csv"
	"io"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs/nvidia_smi/common"
)

const measurement = "ppusmi"

const (
	colIndex = iota
	colName
	colUUID
	colPCIBusID
	colComputeMode
	colMemoryTotal
	colMemoryUsed
	colMemoryFree
	colTemperatureGPU
	colTemperatureMemory
	colUtilizationGPU
	colUtilizationMemory
	colPowerDraw
	colClocksSM
	colClocksMemory
	colCount
)

func parse(acc telegraf.Accumulator, buf []byte) error {
	r := csv.NewReader(bytes.NewReader(buf))
	r.TrimLeadingSpace = true
	r.FieldsPerRecord = -1

	for {
		rec, err := r.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if skipCSVRecord(rec) {
			continue
		}
		if len(rec) < colCount {
			continue
		}

		index := strings.TrimSpace(rec[colIndex])
		if index == "" || strings.EqualFold(index, "N/A") {
			continue
		}

		tags := map[string]string{
			"index": index,
		}
		common.SetTagIfUsed(tags, "name", rec[colName])
		common.SetTagIfUsed(tags, "uuid", rec[colUUID])
		common.SetTagIfUsed(tags, "pci_bus_id", rec[colPCIBusID])
		common.SetTagIfUsed(tags, "compute_mode", rec[colComputeMode])

		fields := make(map[string]interface{}, 10)
		common.SetIfUsed("int", fields, "memory_total", rec[colMemoryTotal])
		common.SetIfUsed("int", fields, "memory_used", rec[colMemoryUsed])
		common.SetIfUsed("int", fields, "memory_free", rec[colMemoryFree])
		common.SetIfUsed("int", fields, "temperature_gpu", rec[colTemperatureGPU])
		common.SetIfUsed("int", fields, "temperature_memory", rec[colTemperatureMemory])
		common.SetIfUsed("int", fields, "utilization_gpu", rec[colUtilizationGPU])
		common.SetIfUsed("int", fields, "utilization_memory", rec[colUtilizationMemory])
		common.SetIfUsed("float", fields, "power_draw", rec[colPowerDraw])
		common.SetIfUsed("int", fields, "clocks_current_sm", rec[colClocksSM])
		common.SetIfUsed("int", fields, "clocks_current_memory", rec[colClocksMemory])
		acc.AddFields(measurement, fields, tags)
	}
}

func skipCSVRecord(rec []string) bool {
	if len(rec) == 0 {
		return true
	}
	first := strings.ToLower(strings.TrimSpace(rec[0]))
	return first == "" || first == "index" || first == "timestamp"
}
