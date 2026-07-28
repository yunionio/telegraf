package npu_smi

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

const (
	usagesOutput = `
        NPU ID                         : 0
        Chip Count                     : 1

        DDR Capacity(MB)               : 0
        DDR Usage Rate(%)              : 0
        DDR Hugepages Total(page)      : 0
        DDR Hugepages Usage Rate(%)    : 0
        HBM Capacity(MB)               : 65536
        HBM Usage Rate(%)              : 5
        Aicore Usage Rate(%)           : 0
        Aivector Usage Rate(%)         : 0
        Aicpu Usage Rate(%)            : 0
        Ctrlcpu Usage Rate(%)          : 6
        DDR Bandwidth Usage Rate(%)    : 0
        HBM Bandwidth Usage Rate(%)    : 0
        NPU Utilization(%)             : 0
        Aicube Usage Rate(%)           : 0
        Chip ID                        : 0
`

	boardOutput = `
        NPU ID                         : 0
        Product Name                   : IT21HMDC_Bin5
        Model                          : NA
        Manufacturer                   : Huawei
        Serial Number                  : 1025C9048791
        Software Version               : 25.5.1
        Firmware Version               : 7.8.0.6.201
        Compatibility                  : OK
        Board ID                       : 0x60
        PCB ID                         : A
        BOM ID                         : 1
        PCIe Bus Info                  : 0000:C1:00.0
        Slot ID                        : 0
        Class ID                       : NA
        PCI Vendor ID                  : 0x19E5
        PCI Device ID                  : 0xD802
        Subsystem Vendor ID            : 0x19E5
        Subsystem Device ID            : 0x3003
        Chip Count                     : 1
`

	tempOutput = `
        NPU ID                         : 0
        Chip Count                     : 1

        NPU Temperature (C)            : 39
        Chip ID                        : 0

        LM75A_TE (C)                   : 36
        LM75B_TE (C)                   : 36
        Chip Name                      : MCU
`

	powerOutput = `
        NPU ID                         : 0
        Chip Count                     : 1

        NPU Real-time Power(W)         : 103.5
        Chip ID                        : 0
`

	listOutput = `
        Total Count                    : 8

        NPU ID                         : 0
        Chip Count                     : 1

        NPU ID                         : 1
        Chip Count                     : 1

        NPU ID                         : 2
        Chip Count                     : 1

        NPU ID                         : 3
        Chip Count                     : 1

        NPU ID                         : 4
        Chip Count                     : 1

        NPU ID                         : 5
        Chip Count                     : 1

        NPU ID                         : 6
        Chip Count                     : 1

        NPU ID                         : 7
        Chip Count                     : 1
`
)

func TestParseNpuIDs(t *testing.T) {
	got, err := parseNpuIDs([]byte(listOutput))
	require.NoError(t, err)
	require.Equal(t, []int{0, 1, 2, 3, 4, 5, 6, 7}, got)
}

func TestParseNpuInfo(t *testing.T) {
	info := &NpuInfo{
		NpuID:   0,
		Devices: make(map[int]*Device),
	}

	require.NoError(t, parseUsages(info, []byte(usagesOutput)))
	require.NoError(t, parseBoard(info, []byte(boardOutput)))
	require.NoError(t, parseTemperature(info, []byte(tempOutput)))
	require.NoError(t, parsePower(info, []byte(powerOutput)))

	require.Len(t, info.Devices, 1)
	dev := info.Devices[0]
	require.NotNil(t, dev)
	require.Equal(t, int64(65536), dev.HBMCapacity)
	require.Equal(t, 5.0, dev.HBMUsageRate)
	require.Equal(t, 6.0, dev.CtrlCpuUsageRate)
	require.Equal(t, 39.0, dev.Temperature)
	require.Equal(t, 103.5, dev.Power)
	require.Equal(t, "0000:C1:00.0", info.PCIeBusInfo)
}

func TestCollectResults(t *testing.T) {
	info := &NpuInfo{
		NpuID:           0,
		ProductName:     "IT21HMDC_Bin5",
		Manufacturer:    "Huawei",
		SerialNumber:    "1025C9048791",
		SoftwareVersion: "25.5.1",
		FirmwareVersion: "7.8.0.6.201",
		PCIeBusInfo:     "0000:C1:00.0",
		SlotID:          "0",
		Devices: map[int]*Device{
			0: {
				NpuID:             0,
				ChipID:            0,
				HBMCapacity:       65536,
				HBMUsageRate:      5,
				CtrlCpuUsageRate:  6,
				Temperature:       39,
				Power:             103.5,
				NpuUtilization:    0,
				AICoreUsageRate:   0,
				AIVectorUsageRate: 0,
			},
		},
	}

	var acc testutil.Accumulator
	collectResults(&acc, []*NpuInfo{info})

	require.Len(t, acc.Metrics, 1)
	metric := acc.Metrics[0]
	require.Equal(t, measurement, metric.Measurement)
	require.Equal(t, "0", metric.Tags["npu_id"])
	require.Equal(t, "0", metric.Tags["chip_id"])
	require.Equal(t, "IT21HMDC_Bin5", metric.Tags["product_name"])
	require.Equal(t, "0000:C1:00.0", metric.Tags["pcie_bus_info"])
	require.Equal(t, int64(65536), metric.Fields["hbm_capacity"])
	require.Equal(t, 5.0, metric.Fields["hbm_usage_rate"])
	require.Equal(t, 39.0, metric.Fields["temperature"])
	require.Equal(t, 103.5, metric.Fields["npu_real_time_power"])
}
