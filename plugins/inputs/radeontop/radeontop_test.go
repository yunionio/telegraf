//go:build linux

package radeontop

import (
	"reflect"
	"testing"
)

func Test_parseRadeontopLine(t *testing.T) {
	tests := []struct {
		line    string
		want    *result
		wantErr bool
	}{
		{
			line: `1731392941.164959: bus 03, gpu 14.17%, ee 0.00%, vgt 7.50%, ta 4.17%, sx 4.17%, sh 7.50%, spi 8.33%, sc 5.83%, pa 2.50%, db 5.83%, cb 5.83%, vram 62.58% 20482.97mb, gtt 15.94% 5218.41mb, mclk 100.00% 1.000ghz, sclk 94.13% 2.405ghz`,
			want: &result{
				Bus:                       "03",
				GPU:                       14.17,
				EventEngine:               0,
				VertexGrouperTesselator:   7.5,
				TextureAddresser:          4.17,
				ShaderExport:              4.17,
				SequencerInstructionCache: 7.5,
				ShaderInterpolator:        8.33,
				ScanConverter:             5.83,
				PrimitiveAssembly:         2.5,
				DepthBlock:                5.83,
				ColorBlock:                5.83,
				VRAM:                      62.58,
				VRAMUsedSizeMB:            20482,
				VRAMSizeMB:                32730,
				GTT:                       15.94,
				GTTUsedSizeMB:             5218,
				GTTSizeMB:                 32737,
				MemoryClock:               100,
				MemoryClockMHz:            1000,
				ShaderClock:               94.13,
				ShaderClockMHz:            2400,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			got, err := parseRadeontopLine(tt.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRadeontopLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseRadeontopLine() got = %v, want %v", got, tt.want)
			}
		})
	}
}
