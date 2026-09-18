package xpusmi

import (
	"encoding/xml"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs/nvidia_smi/common"
)

const measurement = "xpusmi"

func parse(acc telegraf.Accumulator, buf []byte) error {
	var log xpusmiLog
	if err := xml.Unmarshal(buf, &log); err != nil {
		return err
	}
	for i := range log.Xpu {
		xpu := &log.Xpu[i]
		index := xpu.MinorNumber
		if index == "" {
			index = strconv.Itoa(i)
		}
		tags := map[string]string{
			"index": index,
		}
		common.SetTagIfUsed(tags, "name", xpu.ProductName)
		common.SetTagIfUsed(tags, "uuid", xpu.UUID)
		common.SetTagIfUsed(tags, "pci_bus_id", xpu.ID)
		common.SetTagIfUsed(tags, "oam_id", xpu.OamID)

		fields := make(map[string]interface{}, 32)
		common.SetIfUsed("str", fields, "driver_version", log.DriverVersion)
		common.SetIfUsed("str", fields, "cuda_version", log.CUDAVersion)
		common.SetIfUsed("str", fields, "serial", xpu.Serial)
		common.SetIfUsed("str", fields, "product_brand", xpu.ProductBrand)
		common.SetIfUsed("str", fields, "product_architecture", xpu.ProductArchitecture)
		common.SetIfUsed("str", fields, "current_ecc", xpu.EccMode.CurrentEcc)
		common.SetIfUsed("int", fields, "memory_total", xpu.Memory.Total)
		common.SetIfUsed("int", fields, "memory_used", xpu.Memory.Used)
		common.SetIfUsed("int", fields, "memory_free", xpu.Memory.Free)
		common.SetIfUsed("int", fields, "l3_memory_total", xpu.L3.Total)
		common.SetIfUsed("int", fields, "l3_memory_used", xpu.L3.Used)
		common.SetIfUsed("int", fields, "l3_memory_free", xpu.L3.Free)
		common.SetIfUsed("int", fields, "utilization_gpu", xpu.Utilization.XpuUtil)
		common.SetIfUsed("int", fields, "temperature_gpu", xpu.Temp.XpuTemp)
		common.SetIfUsed("float", fields, "power_draw", xpu.Power.PowerDraw)
		common.SetIfUsed("float", fields, "power_limit", xpu.Power.EnforcedPowerLimit)
		common.SetIfUsed("int", fields, "clocks_current_cluster", xpu.Clocks.ClusterClock)
		common.SetIfUsed("int", fields, "clocks_current_cdnn", xpu.Clocks.CdnnClock)
		common.SetIfUsed("int", fields, "pcie_link_gen_current", xpu.PCI.LinkInfo.PCIEGen.CurrentLinkGen)
		common.SetIfUsed("int", fields, "pcie_link_width_current", xpu.PCI.LinkInfo.LinkWidth.CurrentLinkWidth)
		common.SetIfUsed("int", fields, "ecc_errors_dram_correctable", xpu.EccErrors.Volatile.DramCorrectable)
		common.SetIfUsed("int", fields, "ecc_errors_dram_uncorrectable", xpu.EccErrors.Volatile.DramUncorrectable)
		common.SetIfUsed("int", fields, "ecc_errors_dram_correctable_aggregate", xpu.EccErrors.Aggregate.DramCorrectable)
		common.SetIfUsed("int", fields, "ecc_errors_dram_uncorrectable_aggregate", xpu.EccErrors.Aggregate.DramUncorrectable)
		acc.AddFields(measurement, fields, tags)
	}
	return nil
}
