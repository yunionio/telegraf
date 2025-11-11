//go:build linux

package ni_rsrc_mon

import (
	"reflect"
	"testing"
)

const (
	testJson = `**************************************************
2 devices retrieved from current pool at start up
Wed Nov 13 10:57:40 2024 up 00:00:00 v4866rKr2
{
  "decoders": [
        {
                "NUMBER": 2,
                "INDEX": 0,
                "LOAD": 0,
                "MODEL_LOAD": 0,
                "FW_LOAD": 0,
                "INST": 0,
                "MAX_INST": 128,
                "MEM": 0,
                "CRITICAL_MEM": 0,
                "SHARE_MEM": 1,
                "P2P_MEM": 0,
                "DEVICE": "/dev/nvme0n1",
                "L_FL2V": "4.5.0",
                "N_FL2V": "4.5.0",
                "FR": "4866rKr1",
                "N_FR": "4866rKr1",
                "NUMA_NODE": 2,
                "PCIE_ADDR": "0000:86:00.0"
        }
  ],
  "nvmes": [
        {
                "NUMBER": 2,
                "INDEX": 1,
                "LOAD": 0,
                "MODEL_LOAD": 0,
                "FW_LOAD": 68,
                "INST": 0,
                "MAX_INST": 0,
                "MEM": 0,
                "CRITICAL_MEM": 0,
                "SHARE_MEM": 0,
                "P2P_MEM": 0,
                "DEVICE": "/dev/nvme1n1",
                "L_FL2V": "4.5.0",
                "N_FL2V": "4.5.0",
                "FR": "4866rKr1",
                "N_FR": "4866rKr1",
                "NUMA_NODE": 2,
                "PCIE_ADDR": "0000:87:00.0"
        }
  ]
}
**************************************************`
)

func Test_parseResults(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    *Results
		wantErr bool
	}{
		{
			name:    "empty",
			content: ``,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "normal",
			content: testJson,
			want: &Results{
				Decoders: []*Result{
					{
						Number:      2,
						Index:       0,
						Load:        0,
						ModelLoad:   0,
						FwLoad:      0,
						Inst:        0,
						MaxInst:     128,
						Mem:         0,
						CriticalMem: 0,
						ShareMem:    1,
						P2PMem:      0,
						Device:      "/dev/nvme0n1",
						LFl2v:       "4.5.0",
						NFl2v:       "4.5.0",
						Fr:          "4866rKr1",
						NFr:         "4866rKr1",
						NumaNode:    2,
						PCIEAddr:    "0000:86:00.0",
					},
				},
				Nvmes: []*Result{
					{
						Number:      2,
						Index:       1,
						Load:        0,
						ModelLoad:   0,
						FwLoad:      68,
						Inst:        0,
						MaxInst:     0,
						Mem:         0,
						CriticalMem: 0,
						ShareMem:    0,
						P2PMem:      0,
						Device:      "/dev/nvme1n1",
						LFl2v:       "4.5.0",
						NFl2v:       "4.5.0",
						Fr:          "4866rKr1",
						NFr:         "4866rKr1",
						NumaNode:    2,
						PCIEAddr:    "0000:87:00.0",
					},
				},
			},
			wantErr: false,
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
				t.Errorf("parseResults() got = %v, want %v", got, tt.want)
			}
		})
	}
}
