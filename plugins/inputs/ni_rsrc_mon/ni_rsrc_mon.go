//go:generate ../../../tools/readme_config_includer/generator
package ni_rsrc_mon

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/procutils"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

const measurement = "ni_rsrc_mon"

func init() {
	inputs.Add(measurement, func() telegraf.Input {
		return &niRsrcMon{
			BinPath: "/usr/bin/ni_rsrc_mon",
			Timeout: config.Duration(5 * time.Second),
		}
	})
}

type niRsrcMon struct {
	BinPath string
	Timeout config.Duration
}

func (n *niRsrcMon) SampleConfig() string {
	return sampleConfig
}

func (n niRsrcMon) Gather(accumulator telegraf.Accumulator) error {
	if _, err := procutils.IsRemoteFileExist(n.BinPath); err != nil {
		return err
	}
	results, err := n.pollMetrics()
	if err != nil {
		return errors.Wrap(err, "pollMetrics")
	}
	if err := colloectResults(accumulator, results); err != nil {
		return errors.Wrap(err, "collect result")
	}
	return nil
}

func colloectResults(accumulator telegraf.Accumulator, results *Results) error {
	for _, r := range results.GetResults() {
		accumulator.AddFields(measurement, r.getFields(), r.getTags())
	}
	return nil
}

type Result struct {
	Number      int    `json:"NUMBER"`
	Index       int    `json:"INDEX"`
	Load        int    `json:"LOAD"`
	ModelLoad   int    `json:"MODEL_LOAD"`
	FwLoad      int    `json:"FW_LOAD"`
	Inst        int    `json:"INST"`
	MaxInst     int    `json:"MAX_INST"`
	Mem         int    `json:"MEM"`
	CriticalMem int    `json:"CRITICAL_MEM"`
	ShareMem    int    `json:"SHARE_MEM"`
	P2PMem      int    `json:"P2P_MEM"`
	Device      string `json:"DEVICE"`
	LFl2v       string `json:"L_FL2V"`
	NFl2v       string `json:"N_FL2V"`
	Fr          string `json:"FR"`
	NFr         string `json:"N_FR"`
	NumaNode    int    `json:"NUMA_NODE"`
	PCIEAddr    string `json:"PCIE_ADDR"`
}

func (r Result) getFields() map[string]interface{} {
	return map[string]interface{}{
		"load":         r.Load,
		"model_load":   r.ModelLoad,
		"fw_load":      r.FwLoad,
		"inst":         r.Inst,
		"max_inst":     r.MaxInst,
		"mem":          r.Mem,
		"critical_mem": r.CriticalMem,
		"share_mem":    r.ShareMem,
		"p2p_mem":      r.P2PMem,
	}
}

func (r Result) getTags() map[string]string {
	return map[string]string{
		"number":    fmt.Sprintf("%d", r.Number),
		"index":     fmt.Sprintf("%d", r.Index),
		"device":    r.Device,
		"l_fl2v":    r.LFl2v,
		"n_fl2v":    r.NFl2v,
		"fr":        r.Fr,
		"n_fr":      r.NFr,
		"numa_node": fmt.Sprintf("%d", r.NumaNode),
		"pcie_addr": r.PCIEAddr,
	}
}

type Results struct {
	Decoders  []*Result `json:"decoders"`
	Encoders  []*Result `json:"encoders"`
	Uploaders []*Result `json:"uploaders"`
	Scalers   []*Result `json:"scalers"`
	AIs       []*Result `json:"AIs"`
	Nvmes     []*Result `json:"nvmes"`
}

type IResult interface {
	getFields() map[string]interface{}
	getTags() map[string]string
}

type resultW struct {
	*Result
	Type string
}

func (r resultW) getFields() map[string]interface{} {
	return r.Result.getFields()
}

func (r resultW) getTags() map[string]string {
	tags := r.Result.getTags()
	tags["type"] = r.Type
	return tags
}

func newResultW(result *Result, typ string) *resultW {
	return &resultW{
		Result: result,
		Type:   typ,
	}
}

func (rs *Results) GetResults() []IResult {
	results := make([]IResult, 0)
	for _, decoder := range rs.Decoders {
		results = append(results, newResultW(decoder, "decoder"))
	}
	for _, item := range rs.Encoders {
		results = append(results, newResultW(item, "encoder"))
	}
	for _, item := range rs.Uploaders {
		results = append(results, newResultW(item, "uploader"))
	}
	for _, item := range rs.Scalers {
		results = append(results, newResultW(item, "scaler"))
	}
	for _, item := range rs.AIs {
		results = append(results, newResultW(item, "ai"))
	}
	for _, item := range rs.Nvmes {
		results = append(results, newResultW(item, "nvme"))
	}
	return results
}

func (n niRsrcMon) pollMetrics() (*Results, error) {
	//ret, err := internal.CombinedOutputTimeout(
	//	exec.Command(n.BinPath, "-o", "json1"),
	//	time.Duration(n.Timeout))
	ret, err := procutils.NewRemoteCommandAsFarAsPossible(n.BinPath, "-o", "json1").Output()
	if err != nil {
		return nil, errors.Wrapf(err, "out: %s", ret)
	}
	return parseResults(ret)
}

func parseResults(content []byte) (*Results, error) {
	lines := strings.Split(string(content), "\n")
	jsonStartIdx := 0
	jsonEndIdx := 0
	for i, line := range lines {
		if line == "{" {
			jsonStartIdx = i
		}
		if line == "}" {
			jsonEndIdx = i
			break
		}
	}
	if jsonStartIdx >= jsonEndIdx {
		return nil, errors.Errorf("can't found start(%d) end end(%d) index", jsonStartIdx, jsonEndIdx)
	}
	filterLines := lines[jsonStartIdx : jsonEndIdx+1]
	jsonContent := strings.Join(filterLines, "\n")
	results := new(Results)
	if err := json.Unmarshal([]byte(jsonContent), results); err != nil {
		return nil, errors.Wrapf(err, "can't parse json results: %s", jsonContent)
	}
	return results, nil
}
