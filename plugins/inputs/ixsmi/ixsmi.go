//go:generate ../../../tools/readme_config_includer/generator
package ixsmi

import (
	_ "embed"
	"path/filepath"
	"time"

	"github.com/pkg/errors"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/procutils"
	"github.com/influxdata/telegraf/plugins/inputs"
)

func init() {
	inputs.Add("ixsmi", func() telegraf.Input {
		return &ixsmi{
			BinPath: "/usr/local/corex-4.4.0/bin/ixsmi",
			Timeout: config.Duration(5 * time.Second),
		}
	})
}

//go:embed sample.conf
var sampleConfig string

type ixsmi struct {
	BinPath string          `toml:"bin_path"`
	LibPath string          `toml:"lib_path"`
	Timeout config.Duration `toml:"timeout"`
}

func (s *ixsmi) SampleConfig() string {
	return sampleConfig
}

func (s *ixsmi) Description() string {
	return "Pull statistics from Iluvatar ixsmi attached to the host"
}

func (s *ixsmi) libPath() string {
	if s.LibPath != "" {
		return s.LibPath
	}
	if s.BinPath == "" {
		return "/usr/local/corex-4.4.0/lib64"
	}
	return filepath.Join(filepath.Dir(filepath.Dir(s.BinPath)), "lib64")
}

func (s *ixsmi) Gather(acc telegraf.Accumulator) error {
	if s.BinPath == "" {
		s.BinPath = "/usr/local/corex-4.4.0/bin/ixsmi"
	}
	if _, err := procutils.IsRemoteFileExist(s.BinPath); err != nil {
		return err
	}
	data, err := s.pollMetrics()
	if err != nil {
		return errors.Wrap(err, "pollMetrics")
	}
	return parse(acc, data)
}

func (s *ixsmi) pollMetrics() ([]byte, error) {
	ld := "LD_LIBRARY_PATH=" + s.libPath()
	ret, err := procutils.NewRemoteCommandAsFarAsPossible("env", ld, s.BinPath, "-q", "-x").Output()
	if err != nil {
		return nil, errors.Wrapf(err, "%s -q -x: %s", s.BinPath, ret)
	}
	return ret, nil
}
