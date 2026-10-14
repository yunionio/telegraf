package mthreadsgmi

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

func TestParseS5000JSON(t *testing.T) {
	octets, err := os.ReadFile(filepath.Join("testdata", "s5000-gpu0.json"))
	require.NoError(t, err)

	var acc testutil.Accumulator
	require.NoError(t, parse(&acc, octets))

	metrics := acc.GetTelegrafMetrics()
	require.Len(t, metrics, 1)

	expected0 := testutil.MustMetric(
		"mthreads_gmi",
		map[string]string{
			"index":      "0",
			"name":       "MTT S5000",
			"uuid":       "a19f33bc-dc56-4314-a675-7272e8d831fe",
			"pci_bus_id": "00000000:18:00.0",
			"pstate":     "P0",
		},
		map[string]interface{}{
			"utilization_gpu":         100,
			"utilization_memory":      11,
			"utilization_encoder":     0,
			"utilization_decoder":     0,
			"temperature_gpu":         50,
			"memory_total":            81920,
			"memory_used":             9081,
			"memory_free":             72839,
			"power_draw":              427.47,
			"power_limit":             750.0,
			"clocks_current_graphics": 1750,
			"clocks_current_memory":   2500,
			"pcie_link_gen_current":   5,
			"pcie_link_width_current": 16,
		},
		time.Unix(0, 0),
	)
	testutil.RequireMetricsEqual(t, []telegraf.Metric{expected0}, metrics, testutil.IgnoreTime())
}

func TestParseOmitsNA(t *testing.T) {
	buf := []byte(`{"GPU":{"Index":"1","Product Name":"MTT S5000","DeviceType":"Physical","Utilization":{"Gpu":"1%","Memory":"N/A","Encoder":"N/A","Decoder":"0%"},"Temperature":{"GPU Current Temp":"N/A"},"Power Readings":{"Power Draw ":"N/A","Current Power Limit":"750.00W"}}}`)
	var acc testutil.Accumulator
	require.NoError(t, parse(&acc, buf))
	metrics := acc.GetTelegrafMetrics()
	require.Len(t, metrics, 1)
	_, ok := metrics[0].GetField("utilization_memory")
	require.False(t, ok)
	_, ok = metrics[0].GetField("utilization_encoder")
	require.False(t, ok)
	_, ok = metrics[0].GetField("temperature_gpu")
	require.False(t, ok)
	_, ok = metrics[0].GetField("power_draw")
	require.False(t, ok)
	limit, ok := metrics[0].GetField("power_limit")
	require.True(t, ok)
	require.Equal(t, 750.0, limit)
}
