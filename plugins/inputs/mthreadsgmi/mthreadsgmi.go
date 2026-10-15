package mthreadsgmi

import (
	_ "embed"
	"time"

	"github.com/pkg/errors"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/procutils"
	"github.com/influxdata/telegraf/plugins/inputs"
)

func init() {
	inputs.Add("mthreadsgmi", func() telegraf.Input {
		return &mthreadsGmi{
			BinPath: "/usr/bin/mthreads-gmi",
			Timeout: config.Duration(5 * time.Second),
		}
	})
}

//go:embed sample.conf
var sampleConfig string

type mthreadsGmi struct {
	BinPath string          `toml:"bin_path"`
	Timeout config.Duration `toml:"timeout"`
}

func (m *mthreadsGmi) SampleConfig() string {
	return sampleConfig
}

func (m *mthreadsGmi) Description() string {
	return "Pull statistics from Moore Threads mthreads-gmi attached to the host"
}

func (m *mthreadsGmi) Gather(acc telegraf.Accumulator) error {
	if m.BinPath == "" {
		m.BinPath = "/usr/bin/mthreads-gmi"
	}
	if _, err := procutils.IsRemoteFileExist(m.BinPath); err != nil {
		return err
	}
	data, err := m.pollMetrics()
	if err != nil {
		return errors.Wrap(err, "pollMetrics")
	}
	return parse(acc, data)
}

func (m *mthreadsGmi) pollMetrics() ([]byte, error) {
	ret, err := procutils.NewRemoteCommandAsFarAsPossible(m.BinPath, "-q", "--json").Output()
	if err != nil {
		return nil, errors.Wrapf(err, "%s: %s", m.BinPath, ret)
	}
	return ret, nil
}
