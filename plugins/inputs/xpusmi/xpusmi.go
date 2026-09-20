//go:generate ../../../tools/readme_config_includer/generator
package xpusmi

import (
	_ "embed"
	"time"

	"github.com/pkg/errors"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/procutils"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const defaultBinPath = "/usr/local/bin/xpu-smi"

func init() {
	inputs.Add("xpusmi", func() telegraf.Input {
		return &xpusmi{
			BinPath: defaultBinPath,
			Timeout: config.Duration(5 * time.Second),
		}
	})
}

//go:embed sample.conf
var sampleConfig string

type xpusmi struct {
	BinPath string          `toml:"bin_path"`
	LibPath string          `toml:"lib_path"`
	Timeout config.Duration `toml:"timeout"`
}

func (s *xpusmi) SampleConfig() string {
	return sampleConfig
}

func (s *xpusmi) Description() string {
	return "Pull statistics from Kunlunxin xpu-smi attached to the host"
}

func (s *xpusmi) binPath() string {
	if s.BinPath != "" {
		return s.BinPath
	}
	return defaultBinPath
}

func (s *xpusmi) Gather(acc telegraf.Accumulator) error {
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

func (s *xpusmi) pollMetrics() ([]byte, error) {
	bin := s.binPath()
	args := make([]string, 0, 5)
	if s.LibPath != "" {
		args = append(args, "env", "LD_LIBRARY_PATH="+s.LibPath)
	}
	args = append(args, bin, "-q", "-x")
	ret, err := procutils.NewRemoteCommandAsFarAsPossible(args[0], args[1:]...).Output()
	if err != nil {
		return nil, errors.Wrapf(err, "%s -q -x: %s", bin, ret)
	}
	return ret, nil
}
