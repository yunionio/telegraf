package ppusmi

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

func TestParseZW810E(t *testing.T) {
	octets, err := os.ReadFile(filepath.Join("testdata", "zw810e-2ppu.csv"))
	require.NoError(t, err)

	var acc testutil.Accumulator
	require.NoError(t, parse(&acc, octets))

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"ppusmi",
			map[string]string{
				"index":        "0",
				"name":         "PPU-ZW810E",
				"uuid":         "GPU-019e2226-4211-0208-0000-000000ab261d",
				"compute_mode": "Default",
				"pci_bus_id":   "00000000:a8:00.0",
			},
			map[string]interface{}{
				"memory_total":          98304,
				"memory_used":           248,
				"memory_free":           98056,
				"temperature_gpu":       42,
				"temperature_memory":    42,
				"utilization_gpu":       38,
				"utilization_memory":    0,
				"power_draw":            51.99,
				"clocks_current_sm":     200,
				"clocks_current_memory": 1800,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"ppusmi",
			map[string]string{
				"index":        "1",
				"name":         "PPU-ZW810E",
				"uuid":         "GPU-019e2226-84c1-0200-0000-0000c020e701",
				"compute_mode": "Default",
				"pci_bus_id":   "00000000:a7:00.0",
			},
			map[string]interface{}{
				"memory_total":          98304,
				"utilization_gpu":       0,
				"utilization_memory":    0,
				"clocks_current_sm":     1700,
				"clocks_current_memory": 1800,
			},
			time.Unix(0, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestParseSkipsHeader(t *testing.T) {
	octets, err := os.ReadFile(filepath.Join("testdata", "with-header.csv"))
	require.NoError(t, err)

	var acc testutil.Accumulator
	require.NoError(t, parse(&acc, octets))
	require.Len(t, acc.GetTelegrafMetrics(), 1)
	require.Equal(t, "0", acc.GetTelegrafMetrics()[0].Tags()["index"])
}

func TestLibPathDefault(t *testing.T) {
	s := &ppusmi{BinPath: "/usr/local/bin/ppu-smi"}
	require.Equal(t, "/usr/local/PPU_SDK/lib64", s.libPath())
	s.LibPath = "/opt/PPU_SDK/lib"
	require.Equal(t, "/opt/PPU_SDK/lib", s.libPath())
}

func TestBinPathDefault(t *testing.T) {
	s := &ppusmi{}
	require.Equal(t, "/usr/local/bin/ppu-smi", s.binPath())
}
