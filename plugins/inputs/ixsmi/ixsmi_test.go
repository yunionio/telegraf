package ixsmi

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

func TestParseBI150S(t *testing.T) {
	octets, err := os.ReadFile(filepath.Join("testdata", "bi-v150s-2gpu.xml"))
	require.NoError(t, err)

	var acc testutil.Accumulator
	require.NoError(t, parse(&acc, octets))

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"ixsmi",
			map[string]string{
				"index":        "0",
				"name":         "Iluvatar BI-V150S",
				"uuid":         "GPU-5fe225c8-c806-50f5-9132-90c020e40005",
				"pstate":       "P0",
				"compute_mode": "Default",
				"pci_bus_id":   "00000000:26:00.0",
			},
			map[string]interface{}{
				"driver_version":          "4.4.0",
				"cuda_version":            "10.2",
				"serial":                  "2426005680215E",
				"vbios_version":           "2.0.0",
				"current_ecc":             "Enabled",
				"memory_total":            32768,
				"memory_used":             68,
				"memory_free":             32700,
				"utilization_gpu":         0,
				"utilization_memory":      1,
				"utilization_encoder":     0,
				"utilization_decoder":     0,
				"temperature_gpu":         35,
				"temperature_memory":      34,
				"power_draw":              13.0,
				"power_limit":             205.0,
				"board_power_draw":        53.0,
				"board_power_limit":       450.0,
				"clocks_current_sm":       500,
				"clocks_current_memory":   1600,
				"pcie_link_gen_current":   4,
				"pcie_link_width_current": 16,
				"ecc_errors_single_bit":   0,
				"ecc_errors_double_bit":   0,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"ixsmi",
			map[string]string{
				"index":        "1",
				"name":         "Iluvatar BI-V150S",
				"uuid":         "GPU-a1c65686-2223-5310-9cc7-6079e42e3b35",
				"pstate":       "P0",
				"compute_mode": "Default",
				"pci_bus_id":   "00000000:29:00.0",
			},
			map[string]interface{}{
				"driver_version":          "4.4.0",
				"cuda_version":            "10.2",
				"serial":                  "2426005680215E",
				"vbios_version":           "2.0.0",
				"current_ecc":             "Enabled",
				"memory_total":            32768,
				"memory_used":             68,
				"memory_free":             32700,
				"utilization_gpu":         0,
				"utilization_memory":      1,
				"utilization_encoder":     0,
				"utilization_decoder":     0,
				"temperature_gpu":         33,
				"temperature_memory":      33,
				"power_draw":              13.0,
				"power_limit":             205.0,
				"board_power_draw":        53.0,
				"board_power_limit":       450.0,
				"clocks_current_sm":       500,
				"clocks_current_memory":   1600,
				"pcie_link_gen_current":   4,
				"pcie_link_width_current": 16,
				"ecc_errors_single_bit":   0,
				"ecc_errors_double_bit":   0,
			},
			time.Unix(0, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestLibPathDefault(t *testing.T) {
	s := &ixsmi{BinPath: "/usr/local/corex-4.4.0/bin/ixsmi"}
	require.Equal(t, "/usr/local/corex-4.4.0/lib64", s.libPath())
	s.LibPath = "/opt/corex/lib64"
	require.Equal(t, "/opt/corex/lib64", s.libPath())
}
