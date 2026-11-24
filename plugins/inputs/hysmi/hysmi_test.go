package hysmi

import (
	"reflect"
	"strconv"
	"testing"
)

const testDCUContent = `
============================ System Management Interface =============================
======================================================================================
DCU     Temp     AvgPwr     Perf     PwrCap     VRAM%      DCU%      Mode
0       37.0C    82.0W      auto     400.0W     0%         0.0%      Normal
1       38.0C    78.0W      auto     400.0W     0%         0.0%      Normal
2       37.0C    80.0W      auto     400.0W     0%         0.0%      Normal
3       35.0C    81.0W      auto     400.0W     0%         0.0%      Normal
4       36.0C    77.0W      auto     400.0W     0%         0.0%      Normal
5       36.0C    84.0W      auto     400.0W     0%         0.0%      Normal
6       37.0C    77.0W      auto     400.0W     0%         0.0%      Normal
7       38.0C    84.0W      auto     400.0W     0%         0.0%      Normal
======================================================================================
=================================== End of SMI Log ===================================
`

const testDCUMemContent = `
 /opt/hyhal/bin/hy-smi --showmeminfo all

============================ System Management Interface =============================
======================================================================================
DCU[0]          : gtt Total Memory (MiB): 515891
DCU[0]          : gtt Total Used Memory (MiB): 8
DCU[0]          : vis_vram Total Memory (MiB): 65520
DCU[0]          : vis_vram Total Used Memory (MiB): 2
DCU[0]          : vram Total Memory (MiB): 65520
DCU[0]          : vram Total Used Memory (MiB): 2
DCU[1]          : gtt Total Memory (MiB): 515891
DCU[1]          : gtt Total Used Memory (MiB): 8
DCU[1]          : vis_vram Total Memory (MiB): 65520
DCU[1]          : vis_vram Total Used Memory (MiB): 2
DCU[1]          : vram Total Memory (MiB): 65520
DCU[1]          : vram Total Used Memory (MiB): 2
DCU[2]          : gtt Total Memory (MiB): 515891
DCU[2]          : gtt Total Used Memory (MiB): 8
DCU[2]          : vis_vram Total Memory (MiB): 65520
DCU[2]          : vis_vram Total Used Memory (MiB): 2
DCU[2]          : vram Total Memory (MiB): 65520
DCU[2]          : vram Total Used Memory (MiB): 2
DCU[3]          : gtt Total Memory (MiB): 515891
DCU[3]          : gtt Total Used Memory (MiB): 8
DCU[3]          : vis_vram Total Memory (MiB): 65520
DCU[3]          : vis_vram Total Used Memory (MiB): 2
DCU[3]          : vram Total Memory (MiB): 65520
DCU[3]          : vram Total Used Memory (MiB): 2
DCU[4]          : gtt Total Memory (MiB): 515891
DCU[4]          : gtt Total Used Memory (MiB): 8
DCU[4]          : vis_vram Total Memory (MiB): 65520
DCU[4]          : vis_vram Total Used Memory (MiB): 2
DCU[4]          : vram Total Memory (MiB): 65520
DCU[4]          : vram Total Used Memory (MiB): 2
DCU[5]          : gtt Total Memory (MiB): 515891
DCU[5]          : gtt Total Used Memory (MiB): 8
DCU[5]          : vis_vram Total Memory (MiB): 65520
DCU[5]          : vis_vram Total Used Memory (MiB): 2
DCU[5]          : vram Total Memory (MiB): 65520
DCU[5]          : vram Total Used Memory (MiB): 2
DCU[6]          : gtt Total Memory (MiB): 515891
DCU[6]          : gtt Total Used Memory (MiB): 8
DCU[6]          : vis_vram Total Memory (MiB): 65520
DCU[6]          : vis_vram Total Used Memory (MiB): 2
DCU[6]          : vram Total Memory (MiB): 65520
DCU[6]          : vram Total Used Memory (MiB): 2
DCU[7]          : gtt Total Memory (MiB): 515891
DCU[7]          : gtt Total Used Memory (MiB): 8
DCU[7]          : vis_vram Total Memory (MiB): 65520
DCU[7]          : vis_vram Total Used Memory (MiB): 2
DCU[7]          : vram Total Memory (MiB): 65520
DCU[7]          : vram Total Used Memory (MiB): 2
======================================================================================
=================================== End of SMI Log ===================================
`

const testContent = `
================================= System Management Interface ==================================
================================================================================================
HCU     Temp     AvgPwr     Perf     PwrCap     VRAM%      HCU%      Dec%      Enc%      Mode
0       51.0C    95.0W      auto     1000.0W    0%         0.0%      0.0%      0.0%      Normal
1       50.0C    90.0W      auto     1000.0W    0%         0.0%      0.0%      0.0%      Normal
2       51.0C    89.0W      auto     1000.0W    0%         0.0%      0.0%      0.0%      Normal
3       49.0C    90.0W      auto     1000.0W    0%         0.0%      0.0%      0.0%      Normal
4       52.0C    90.0W      auto     1000.0W    0%         0.0%      0.0%      0.0%      Normal
5       53.0C    91.0W      auto     1000.0W    0%         0.0%      0.0%      0.0%      Normal
6       52.0C    93.0W      auto     1000.0W    0%         0.0%      0.0%      0.0%      Normal
7       53.0C    82.0W      auto     1000.0W    0%         0.0%      0.0%      0.0%      Normal
================================================================================================
======================================== End of SMI Log ========================================
`

const testMemContent = `
================================= System Management Interface ==================================
================================================================================================
HCU[0]          : gtt Total Memory (MiB): 515905
HCU[0]          : gtt Total Used Memory (MiB): 8
HCU[0]          : vis_vram Total Memory (MiB): 65520
HCU[0]          : vis_vram Total Used Memory (MiB): 2
HCU[0]          : vram Total Memory (MiB): 65520
HCU[0]          : vram Total Used Memory (MiB): 2
HCU[1]          : gtt Total Memory (MiB): 515905
HCU[1]          : gtt Total Used Memory (MiB): 16
HCU[1]          : vis_vram Total Memory (MiB): 65520
HCU[1]          : vis_vram Total Used Memory (MiB): 10
HCU[1]          : vram Total Memory (MiB): 65520
HCU[1]          : vram Total Used Memory (MiB): 10
================================================================================================
======================================== End of SMI Log ========================================
`

func Test_parseResults(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []*Hysmi
		wantErr bool
	}{
		{
			name:    "normal",
			content: testContent,
			want: []*Hysmi{
				{Index: "0", Temp: 51, AvgPwr: 95, Perf: "auto", PwrCap: 1000, VRAM: 0, Util: 0, Dec: 0, Enc: 0, Mode: "Normal"},
				{Index: "1", Temp: 50, AvgPwr: 90, Perf: "auto", PwrCap: 1000, VRAM: 0, Util: 0, Dec: 0, Enc: 0, Mode: "Normal"},
				{Index: "2", Temp: 51, AvgPwr: 89, Perf: "auto", PwrCap: 1000, VRAM: 0, Util: 0, Dec: 0, Enc: 0, Mode: "Normal"},
				{Index: "3", Temp: 49, AvgPwr: 90, Perf: "auto", PwrCap: 1000, VRAM: 0, Util: 0, Dec: 0, Enc: 0, Mode: "Normal"},
				{Index: "4", Temp: 52, AvgPwr: 90, Perf: "auto", PwrCap: 1000, VRAM: 0, Util: 0, Dec: 0, Enc: 0, Mode: "Normal"},
				{Index: "5", Temp: 53, AvgPwr: 91, Perf: "auto", PwrCap: 1000, VRAM: 0, Util: 0, Dec: 0, Enc: 0, Mode: "Normal"},
				{Index: "6", Temp: 52, AvgPwr: 93, Perf: "auto", PwrCap: 1000, VRAM: 0, Util: 0, Dec: 0, Enc: 0, Mode: "Normal"},
				{Index: "7", Temp: 53, AvgPwr: 82, Perf: "auto", PwrCap: 1000, VRAM: 0, Util: 0, Dec: 0, Enc: 0, Mode: "Normal"},
			},
		},
		{
			name:    "dcu",
			content: testDCUContent,
			want: []*Hysmi{
				{Index: "0", Temp: 37, AvgPwr: 82, Perf: "auto", PwrCap: 400, VRAM: 0, Util: 0, Mode: "Normal"},
				{Index: "1", Temp: 38, AvgPwr: 78, Perf: "auto", PwrCap: 400, VRAM: 0, Util: 0, Mode: "Normal"},
				{Index: "2", Temp: 37, AvgPwr: 80, Perf: "auto", PwrCap: 400, VRAM: 0, Util: 0, Mode: "Normal"},
				{Index: "3", Temp: 35, AvgPwr: 81, Perf: "auto", PwrCap: 400, VRAM: 0, Util: 0, Mode: "Normal"},
				{Index: "4", Temp: 36, AvgPwr: 77, Perf: "auto", PwrCap: 400, VRAM: 0, Util: 0, Mode: "Normal"},
				{Index: "5", Temp: 36, AvgPwr: 84, Perf: "auto", PwrCap: 400, VRAM: 0, Util: 0, Mode: "Normal"},
				{Index: "6", Temp: 37, AvgPwr: 77, Perf: "auto", PwrCap: 400, VRAM: 0, Util: 0, Mode: "Normal"},
				{Index: "7", Temp: 38, AvgPwr: 84, Perf: "auto", PwrCap: 400, VRAM: 0, Util: 0, Mode: "Normal"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := parseResults([]byte(tt.content))
			if (err != nil) != tt.wantErr {
				t.Errorf("parseResults() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseResults() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Test_parseDCUMemoryInfo(t *testing.T) {
	dcus := make([]*Hysmi, 8)
	for i := range dcus {
		dcus[i] = &Hysmi{Index: strconv.Itoa(i)}
	}
	if err := parseMemoryInfo(dcus, []byte(testDCUMemContent), "DCU["); err != nil {
		t.Fatalf("parseMemoryInfo() error = %v", err)
	}
	if dcus[0].GttTotal != 515891 || dcus[0].GttUsed != 8 || dcus[0].GttFree != 515883 {
		t.Errorf("dcu 0 gtt = %d/%d/%d, want 515891/8/515883", dcus[0].GttTotal, dcus[0].GttUsed, dcus[0].GttFree)
	}
	if dcus[7].VisVramTotal != 65520 || dcus[7].VisVramUsed != 2 || dcus[7].VisVramFree != 65518 {
		t.Errorf("dcu 7 vis_vram = %d/%d/%d, want 65520/2/65518", dcus[7].VisVramTotal, dcus[7].VisVramUsed, dcus[7].VisVramFree)
	}
	if dcus[7].VramTotal != 65520 || dcus[7].VramUsed != 2 || dcus[7].VramFree != 65518 {
		t.Errorf("dcu 7 vram = %d/%d/%d, want 65520/2/65518", dcus[7].VramTotal, dcus[7].VramUsed, dcus[7].VramFree)
	}
}

func Test_parseMemoryInfo(t *testing.T) {
	hcus := []*Hysmi{
		{Index: "0"},
		{Index: "1"},
	}
	if err := parseMemoryInfo(hcus, []byte(testMemContent), "HCU["); err != nil {
		t.Fatalf("parseMemoryInfo() error = %v", err)
	}
	if hcus[0].GttTotal != 515905 || hcus[0].GttUsed != 8 || hcus[0].GttFree != 515897 {
		t.Errorf("hcu 0 gtt = %d/%d/%d, want 515905/8/515897", hcus[0].GttTotal, hcus[0].GttUsed, hcus[0].GttFree)
	}
	if hcus[0].VisVramTotal != 65520 || hcus[0].VisVramUsed != 2 || hcus[0].VisVramFree != 65518 {
		t.Errorf("hcu 0 vis_vram = %d/%d/%d, want 65520/2/65518", hcus[0].VisVramTotal, hcus[0].VisVramUsed, hcus[0].VisVramFree)
	}
	if hcus[0].VramTotal != 65520 || hcus[0].VramUsed != 2 || hcus[0].VramFree != 65518 {
		t.Errorf("hcu 0 vram = %d/%d/%d, want 65520/2/65518", hcus[0].VramTotal, hcus[0].VramUsed, hcus[0].VramFree)
	}
	if hcus[1].GttTotal != 515905 || hcus[1].GttUsed != 16 || hcus[1].GttFree != 515889 {
		t.Errorf("hcu 1 gtt = %d/%d/%d, want 515905/16/515889", hcus[1].GttTotal, hcus[1].GttUsed, hcus[1].GttFree)
	}

	// Unknown HCU index and malformed lines should be ignored.
	badContent := `
HCU[99]         : vram Total Memory (MiB): 100
HCU[abc]        : vram Total Memory (MiB): 100
HCU[0]          : unknown Total Memory (MiB): 100
HCU[0]          : vram Unknown Memory (MiB): 100
HCU[0]          : vram Total Memory (MiB): not_a_number
`
	if err := parseMemoryInfo(hcus, []byte(badContent), "HCU["); err != nil {
		t.Fatalf("parseMemoryInfo() error = %v", err)
	}
	// hcu 0 values should remain unchanged.
	if hcus[0].VramTotal != 65520 {
		t.Errorf("hcu 0 vram total changed to %d", hcus[0].VramTotal)
	}
}

func Test_HCU_getFieldsAndTags(t *testing.T) {
	hcu := &Hysmi{
		Index: "0", Temp: 51, AvgPwr: 95, Perf: "auto", PwrCap: 1000,
		VRAM: 10.5, Util: 20.3, Dec: 1.2, Enc: 3.4, Mode: "Normal",
		VramTotal: 65520, VramUsed: 2, VramFree: 65518,
		GttTotal: 515905, GttUsed: 8, GttFree: 515897,
		VisVramTotal: 65520, VisVramUsed: 2, VisVramFree: 65518,
	}
	fields := hcu.getFields()
	if fields["temperature_gpu"] != 51.0 {
		t.Errorf("temperature_gpu = %v, want 51", fields["temperature_gpu"])
	}
	if fields["power_draw"] != 95.0 {
		t.Errorf("power_draw = %v, want 95", fields["power_draw"])
	}
	if fields["utilization_memory"] != 10.5 {
		t.Errorf("utilization_memory = %v, want 10.5", fields["utilization_memory"])
	}
	if fields["memory_total"] != 65520 {
		t.Errorf("memory_total = %v, want 65520", fields["memory_total"])
	}
	if fields["memory_used"] != 2 {
		t.Errorf("memory_used = %v, want 2", fields["memory_used"])
	}
	if fields["memory_free"] != 65518 {
		t.Errorf("memory_free = %v, want 65518", fields["memory_free"])
	}
	if fields["memory_gtt_total"] != 515905 {
		t.Errorf("memory_gtt_total = %v, want 515905", fields["memory_gtt_total"])
	}
	if fields["memory_vis_vram_total"] != 65520 {
		t.Errorf("memory_vis_vram_total = %v, want 65520", fields["memory_vis_vram_total"])
	}
	tags := hcu.getTags()
	if tags["hcu"] != "0" || tags["perf_mode"] != "auto" || tags["mode"] != "Normal" {
		t.Errorf("getTags() = %v", tags)
	}
}
