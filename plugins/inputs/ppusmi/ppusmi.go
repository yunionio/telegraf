//go:generate ../../../tools/readme_config_includer/generator
package ppusmi

import (
	_ "embed"
	"time"

	"github.com/pkg/errors"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/procutils"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	defaultBinPath = "/usr/local/bin/ppu-smi"
	defaultLibPath = "/usr/local/PPU_SDK/lib64"
	queryFormat    = "csv,noheader,nounits"
)

// nvidia-smi compatible property names (PPU-SMI is a clone).
const queryPPUFields = "index,name,uuid,pci.bus_id,compute_mode,memory.total,memory.used,memory.free,temperature.gpu,temperature.memory,utilization.gpu,utilization.memory,power.draw,clocks.current.sm,clocks.current.memory"

// Aliases from ppu-smi --help-query-ppu / CLOCK (CU) / UTILIZATION (Ppu).
const queryPPUFieldsFallback = "index,name,uuid,pci.bus_id,compute_mode,memory.total,memory.used,memory.free,temperature.ppu,temperature.memory,utilization.ppu,utilization.memory,power.draw,clocks.current.cu,clocks.current.memory"

func init() {
	inputs.Add("ppusmi", func() telegraf.Input {
		return &ppusmi{
			BinPath: defaultBinPath,
			LibPath: defaultLibPath,
			Timeout: config.Duration(5 * time.Second),
		}
	})
}

//go:embed sample.conf
var sampleConfig string

type ppusmi struct {
	BinPath string          `toml:"bin_path"`
	LibPath string          `toml:"lib_path"`
	Timeout config.Duration `toml:"timeout"`
}

func (s *ppusmi) SampleConfig() string {
	return sampleConfig
}

func (s *ppusmi) Description() string {
	return "Pull statistics from T-Head ppu-smi attached to the host"
}

func (s *ppusmi) libPath() string {
	if s.LibPath != "" {
		return s.LibPath
	}
	return defaultLibPath
}

func (s *ppusmi) binPath() string {
	if s.BinPath != "" {
		return s.BinPath
	}
	return defaultBinPath
}

func (s *ppusmi) Gather(acc telegraf.Accumulator) error {
	bin := s.binPath()
	if _, err := procutils.IsRemoteFileExist(bin); err != nil {
		return err
	}
	data, err := s.pollMetrics()
	if err != nil {
		return errors.Wrap(err, "pollMetrics")
	}
	return parse(acc, data)
}

func (s *ppusmi) pollMetrics() ([]byte, error) {
	ret, err := s.runQuery(queryPPUFields)
	if err == nil {
		return ret, nil
	}
	fallback, ferr := s.runQuery(queryPPUFieldsFallback)
	if ferr != nil {
		return nil, errors.Wrapf(err, "ppu-smi query (fallback also failed: %v)", ferr)
	}
	return fallback, nil
}

func (s *ppusmi) runQuery(fields string) ([]byte, error) {
	ld := "LD_LIBRARY_PATH=" + s.libPath()
	bin := s.binPath()
	ret, err := procutils.NewRemoteCommandAsFarAsPossible("env", ld, bin, "--query-ppu="+fields, "--format="+queryFormat).Output()
	if err != nil {
		return nil, errors.Wrapf(err, "%s --query-ppu: %s", bin, ret)
	}
	return ret, nil
}
