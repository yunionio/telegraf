package ixsmi

import (
	"encoding/xml"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs/nvidia_smi/common"
)

const measurement = "ixsmi"

func parse(acc telegraf.Accumulator, buf []byte) error {
	var log ixsmiLog
	if err := xml.Unmarshal(buf, &log); err != nil {
		return err
	}
	for i := range log.GPU {
		gpu := &log.GPU[i]
		index := gpu.MinorNumber
		if index == "" {
			index = strconv.Itoa(i)
		}
		tags := map[string]string{
			"index": index,
		}
		common.SetTagIfUsed(tags, "pstate", gpu.PState)
		common.SetTagIfUsed(tags, "name", gpu.ProdName)
		common.SetTagIfUsed(tags, "uuid", gpu.UUID)
		common.SetTagIfUsed(tags, "compute_mode", gpu.ComputeMode)
		common.SetTagIfUsed(tags, "pci_bus_id", gpu.ID)

		fields := make(map[string]interface{}, 32)
		common.SetIfUsed("str", fields, "driver_version", log.DriverVersion)
		common.SetIfUsed("str", fields, "cuda_version", log.CUDAVersion)
		common.SetIfUsed("str", fields, "serial", gpu.Serial)
		common.SetIfUsed("str", fields, "vbios_version", gpu.VbiosVersion)
		common.SetIfUsed("str", fields, "current_ecc", gpu.EccMode.CurrentEcc)
		common.SetIfUsed("int", fields, "fan_speed", gpu.FanSpeed)
		common.SetIfUsed("int", fields, "memory_total", gpu.Memory.Total)
		common.SetIfUsed("int", fields, "memory_used", gpu.Memory.Used)
		common.SetIfUsed("int", fields, "memory_free", gpu.Memory.Free)
		common.SetIfUsed("int", fields, "utilization_gpu", gpu.Utilization.GPU)
		common.SetIfUsed("int", fields, "utilization_memory", gpu.Utilization.Memory)
		common.SetIfUsed("int", fields, "utilization_encoder", gpu.Utilization.Encoder)
		common.SetIfUsed("int", fields, "utilization_decoder", gpu.Utilization.Decoder)
		common.SetIfUsed("int", fields, "temperature_gpu", gpu.Temp.GPUTemp)
		common.SetIfUsed("int", fields, "temperature_memory", gpu.Temp.MemoryTemp)
		common.SetIfUsed("float", fields, "power_draw", gpu.Power.GPUPowerDraw)
		common.SetIfUsed("float", fields, "power_limit", gpu.Power.CurrentGPUPowerLimit)
		common.SetIfUsed("float", fields, "board_power_draw", gpu.Power.BoardPowerDraw)
		common.SetIfUsed("float", fields, "board_power_limit", gpu.Power.CurrentBoardPowerLimit)
		common.SetIfUsed("int", fields, "clocks_current_sm", gpu.Clocks.SM)
		common.SetIfUsed("int", fields, "clocks_current_memory", gpu.Clocks.Memory)
		common.SetIfUsed("int", fields, "pcie_link_gen_current", gpu.PCI.LinkInfo.PCIEGen.CurrentLinkGen)
		common.SetIfUsed("int", fields, "pcie_link_width_current", gpu.PCI.LinkInfo.LinkWidth.CurrentLinkWidth)
		common.SetIfUsed("int", fields, "ecc_errors_single_bit", gpu.EccErrors.SingleBit)
		common.SetIfUsed("int", fields, "ecc_errors_double_bit", gpu.EccErrors.DoubleBit)
		acc.AddFields(measurement, fields, tags)
	}
	return nil
}
