package xpusmi

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

func TestParseP800OAM(t *testing.T) {
	octets, err := os.ReadFile(filepath.Join("testdata", "p800-oam-8xpu.xml"))
	require.NoError(t, err)

	var acc testutil.Accumulator
	require.NoError(t, parse(&acc, octets))

	metrics := acc.GetTelegrafMetrics()
	require.Len(t, metrics, 8)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"xpusmi",
			map[string]string{
				"index":      "0",
				"name":       "P800 OAM",
				"uuid":       "GPU-a0ac042a-5a14-5481-bf7e-dd8e523792cd",
				"pci_bus_id": "00000000:03:00.0",
				"oam_id":     "3",
			},
			map[string]interface{}{
				"driver_version":                          "5.19.0.0",
				"cuda_version":                            "10.2",
				"serial":                                  "02K15K6252V00115",
				"product_brand":                           "KUNLUNXIN",
				"product_architecture":                    "KL3",
				"current_ecc":                             "Enabled",
				"memory_total":                            98304,
				"memory_used":                             0,
				"memory_free":                             98304,
				"l3_memory_total":                         96,
				"l3_memory_used":                          0,
				"l3_memory_free":                          96,
				"utilization_gpu":                         0,
				"temperature_gpu":                         36,
				"power_draw":                              87.0,
				"power_limit":                             400.0,
				"clocks_current_cluster":                  1450,
				"clocks_current_cdnn":                     1450,
				"pcie_link_gen_current":                   4,
				"pcie_link_width_current":                 16,
				"ecc_errors_dram_correctable":             0,
				"ecc_errors_dram_uncorrectable":           0,
				"ecc_errors_dram_correctable_aggregate":   0,
				"ecc_errors_dram_uncorrectable_aggregate": 0,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"xpusmi",
			map[string]string{
				"index":      "7",
				"name":       "P800 OAM",
				"uuid":       "GPU-5e4fa5ff-6c8b-5873-ab72-c3e5a176da72",
				"pci_bus_id": "00000000:A5:00.0",
				"oam_id":     "5",
			},
			map[string]interface{}{
				"driver_version":                          "5.19.0.0",
				"cuda_version":                            "10.2",
				"serial":                                  "02K15K6252V0007V",
				"product_brand":                           "KUNLUNXIN",
				"product_architecture":                    "KL3",
				"current_ecc":                             "Enabled",
				"memory_total":                            98304,
				"memory_used":                             0,
				"memory_free":                             98304,
				"l3_memory_total":                         96,
				"l3_memory_used":                          0,
				"l3_memory_free":                          96,
				"utilization_gpu":                         0,
				"temperature_gpu":                         36,
				"power_draw":                              85.0,
				"power_limit":                             400.0,
				"clocks_current_cluster":                  1450,
				"clocks_current_cdnn":                     1450,
				"pcie_link_gen_current":                   4,
				"pcie_link_width_current":                 16,
				"ecc_errors_dram_correctable":             0,
				"ecc_errors_dram_uncorrectable":           0,
				"ecc_errors_dram_correctable_aggregate":   0,
				"ecc_errors_dram_uncorrectable_aggregate": 0,
			},
			time.Unix(0, 0),
		),
	}
	testutil.RequireMetricsSubset(t, expected, metrics, testutil.IgnoreTime())
}

func TestParseOmitsInvalidValues(t *testing.T) {
	input := []byte(`<?xml version="1.0" ?>
<!DOCTYPE xpu_smi_log SYSTEM "xpusmi_device_v11.dtd">
<xpu_smi_log>
        <driver_version>5.19.0.0</driver_version>
        <attached_xpus>1</attached_xpus>
        <xpu id="00000000:03:00.0">
                <product_name>P800 OAM</product_name>
                <minor_number>0</minor_number>
                <fb_memory_usage>
                        <total>98304 MiB</total>
                        <used>N/A</used>
                        <free>98304 MiB</free>
                </fb_memory_usage>
                <utilization>
                        <xpu_util>Invalid Argument</xpu_util>
                </utilization>
                <temperature>
                        <xpu_temp>36 C</xpu_temp>
                </temperature>
                <power_readings>
                        <enforced_power_limit>400.00 W</enforced_power_limit>
                        <power_draw>N/A</power_draw>
                </power_readings>
        </xpu>
</xpu_smi_log>`)

	var acc testutil.Accumulator
	require.NoError(t, parse(&acc, input))

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"xpusmi",
			map[string]string{
				"index":      "0",
				"name":       "P800 OAM",
				"pci_bus_id": "00000000:03:00.0",
			},
			map[string]interface{}{
				"driver_version":  "5.19.0.0",
				"memory_total":    98304,
				"memory_free":     98304,
				"temperature_gpu": 36,
				"power_limit":     400.0,
			},
			time.Unix(0, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestIndexFallsBackToArrayOrder(t *testing.T) {
	input := []byte(`<?xml version="1.0" ?>
<xpu_smi_log>
        <attached_xpus>2</attached_xpus>
        <xpu id="00000000:03:00.0">
                <product_name>P800 OAM</product_name>
                <temperature>
                        <xpu_temp>36 C</xpu_temp>
                </temperature>
        </xpu>
        <xpu id="00000000:05:00.0">
                <product_name>P800 OAM</product_name>
                <minor_number>3</minor_number>
                <temperature>
                        <xpu_temp>37 C</xpu_temp>
                </temperature>
        </xpu>
</xpu_smi_log>`)

	var acc testutil.Accumulator
	require.NoError(t, parse(&acc, input))

	metrics := acc.GetTelegrafMetrics()
	require.Len(t, metrics, 2)
	require.Equal(t, "0", metrics[0].Tags()["index"])
	require.Equal(t, "3", metrics[1].Tags()["index"])
}

func TestBinPathDefault(t *testing.T) {
	s := &xpusmi{}
	require.Equal(t, defaultBinPath, s.binPath())
	s.BinPath = "/opt/xpu/bin/xpu-smi"
	require.Equal(t, "/opt/xpu/bin/xpu-smi", s.binPath())
}
