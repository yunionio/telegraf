package brsmi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	// brsmi gpu  dmon -c 1
	testContent = `
 #  gpu    pwr  gtemp   mtemp    spc    mem    enc    dec   mclk   pclk
 #  Idx      W      C       C      %      %      %      %    MHz    MHz
      0    118     42    0~85      0      0      0      0   1800   1100
      1    118     37    0~85      0      0      0      0   1800   1100
      2    118     37    0~85      0      0      0      0   1800   1100
      3    117     41    0~85      0      0      0      0   1800   1100
      4    118     43    0~85      0      0      0      0   1800   1100
      5    119     36    0~85      0      0      0      0   1800   1100
      6    116     37    0~85      0      0      0      0   1800   1100
      7    118     42    0~85      0      0      0      0   1800   1100
`
)

func TestParseResults(t *testing.T) {
	got, err := parseResults([]byte(testContent))
	require.NoError(t, err)
	require.Len(t, got, 8)
	require.Equal(t, &GPU{Index: 0, Power: 118, GTemp: 42, MTemp: 0, MClk: 1800, PClk: 1100}, got[0])
	require.Equal(t, &GPU{Index: 7, Power: 118, GTemp: 42, MTemp: 0, MClk: 1800, PClk: 1100}, got[7])
}

func TestParseResultsInvalid(t *testing.T) {
	_, err := parseResults([]byte("# gpu data\n0 118 bad 0~85 0 0 0 0 1800 1100\n"))
	require.Error(t, err)
}
