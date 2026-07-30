package hysmi

import (
	"reflect"
	"testing"
)

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

func Test_parseResults(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []*HCU
		wantErr bool
	}{
		{
			name:    "normal",
			content: testContent,
			want: []*HCU{
				{Index: "0", Temp: 51, AvgPwr: 95, Perf: "auto", PwrCap: 1000, VRAM: 0, HCUUtil: 0, Dec: 0, Enc: 0, Mode: "Normal"},
				{Index: "1", Temp: 50, AvgPwr: 90, Perf: "auto", PwrCap: 1000, VRAM: 0, HCUUtil: 0, Dec: 0, Enc: 0, Mode: "Normal"},
				{Index: "2", Temp: 51, AvgPwr: 89, Perf: "auto", PwrCap: 1000, VRAM: 0, HCUUtil: 0, Dec: 0, Enc: 0, Mode: "Normal"},
				{Index: "3", Temp: 49, AvgPwr: 90, Perf: "auto", PwrCap: 1000, VRAM: 0, HCUUtil: 0, Dec: 0, Enc: 0, Mode: "Normal"},
				{Index: "4", Temp: 52, AvgPwr: 90, Perf: "auto", PwrCap: 1000, VRAM: 0, HCUUtil: 0, Dec: 0, Enc: 0, Mode: "Normal"},
				{Index: "5", Temp: 53, AvgPwr: 91, Perf: "auto", PwrCap: 1000, VRAM: 0, HCUUtil: 0, Dec: 0, Enc: 0, Mode: "Normal"},
				{Index: "6", Temp: 52, AvgPwr: 93, Perf: "auto", PwrCap: 1000, VRAM: 0, HCUUtil: 0, Dec: 0, Enc: 0, Mode: "Normal"},
				{Index: "7", Temp: 53, AvgPwr: 82, Perf: "auto", PwrCap: 1000, VRAM: 0, HCUUtil: 0, Dec: 0, Enc: 0, Mode: "Normal"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseResults([]byte(tt.content))
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

func Test_HCU_getFieldsAndTags(t *testing.T) {
	hcu := &HCU{
		Index: "0", Temp: 51, AvgPwr: 95, Perf: "auto", PwrCap: 1000,
		VRAM: 10.5, HCUUtil: 20.3, Dec: 1.2, Enc: 3.4, Mode: "Normal",
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
	tags := hcu.getTags()
	if tags["hcu"] != "0" || tags["perf_mode"] != "auto" || tags["mode"] != "Normal" {
		t.Errorf("getTags() = %v", tags)
	}
}
