package vasmi

import (
	"reflect"
	"testing"
)

const (
	testContent = `Smi version:3.1.3
^[[1;1H^[[2JDevice Monitor of AIC
AIC      Pwr |  DevId   DieId    Temp   Oclk   Dclk   Eclk   Gclk  Share_Mem    Gpu    Dec    Enc    AI  Gfx_Mem   Dsp
           W |                      C    MHz    MHz    MHz    MHz          %      %      %      %     %        %     %
-------------*--------------------------------------------------------------------------------------------------------
  0     14.9 |      0       0    61.9     20   1100   1050   1550      75.00   3.86   0.00   2.47  0.00    19.88   N/A
                    1       0    61.6     20   1100   1050   1550      75.00   0.00   0.00   1.94  0.00    19.26   N/A
  1     14.3 |      4       0    59.6     20   1100   1050   1550      85.94   6.92   0.00   4.71  0.00    27.84   N/A
                    7       0    59.3     20   1100   1050   1550      75.00   3.96   0.67  19.03  0.00    26.80   N/A
<1/1>`
)

func Test_parseResults(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []*AIC
		wantErr bool
	}{
		{
			name:    "normal",
			content: testContent,
			want: []*AIC{
				{
					AIC:   "0",
					Power: 14.9,
					Devices: []*Device{
						{
							DevId:    "0",
							DieId:    "0",
							Temp:     61.9,
							Oclk:     20,
							Dclk:     1100,
							Eclk:     1050,
							Gclk:     1550,
							ShareMem: 75,
							GPU:      3.86,
							Dec:      0,
							Enc:      2.47,
							AI:       0,
							GfxMem:   19.88,
						},
						{
							DevId:    "1",
							DieId:    "0",
							Temp:     61.6,
							Oclk:     20,
							Dclk:     1100,
							Eclk:     1050,
							Gclk:     1550,
							ShareMem: 75,
							GPU:      0,
							Dec:      0,
							Enc:      1.94,
							AI:       0,
							GfxMem:   19.26,
						},
					},
				},
				{
					AIC:   "1",
					Power: 14.3,
					Devices: []*Device{
						{
							DevId:    "4",
							DieId:    "0",
							Temp:     59.6,
							Oclk:     20,
							Dclk:     1100,
							Eclk:     1050,
							Gclk:     1550,
							ShareMem: 85.94,
							GPU:      6.92,
							Dec:      0,
							Enc:      4.71,
							AI:       0,
							GfxMem:   27.84,
						},
						{
							DevId:    "7",
							DieId:    "0",
							Temp:     59.3,
							Oclk:     20,
							Dclk:     1100,
							Eclk:     1050,
							Gclk:     1550,
							ShareMem: 75,
							GPU:      3.96,
							Dec:      0.67,
							Enc:      19.03,
							AI:       0,
							GfxMem:   26.8,
						},
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
				t.Errorf("parseResults() got = %#v, want %#v", got, tt.want)
			}
		})
	}
}
