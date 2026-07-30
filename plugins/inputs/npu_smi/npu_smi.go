package npu_smi

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	measurement = "npu_smi"
	binaryName  = "npu-smi"
)

var execCommand = exec.Command // execCommand is used to mock commands in tests.

func init() {
	inputs.Add(measurement, func() telegraf.Input {
		return &NpuSMI{
			BinPath: "/usr/local/bin/npu-smi",
			Timeout: config.Duration(5 * time.Second),
		}
	})
}

type NpuSMI struct {
	BinPath string          `toml:"bin_path"`
	Timeout config.Duration `toml:"timeout"`
}

func (*NpuSMI) SampleConfig() string {
	return `
  ## Optional: path to npu-smi binary.
  # bin_path = "/usr/local/bin/npu-smi"

  ## Optional: timeout for NPU polling.
  # timeout = "5s"
`
}

func (*NpuSMI) Description() string {
	return "Pull statistics from npu-smi info attached to the host"
}

func (smi *NpuSMI) Init() error {
	if smi.BinPath != "" {
		return nil
	}

	binPath, err := exec.LookPath(binaryName)
	if err != nil {
		return fmt.Errorf("looking up %q failed: %w", binaryName, err)
	}
	smi.BinPath = binPath
	return nil
}

func (smi *NpuSMI) Gather(acc telegraf.Accumulator) error {
	binPath := smi.binPath()
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		return fmt.Errorf("npu-smi binary not at path %s, cannot gather NPU data", smi.BinPath)
	}

	results, err := smi.pollMetrics()
	if err != nil {
		return fmt.Errorf("polling metrics failed: %w", err)
	}
	collectResults(acc, results)
	return nil
}

func (smi *NpuSMI) binPath() string {
	hostMountPrefix := os.Getenv("HOST_MOUNT_PREFIX")
	if hostMountPrefix == "" {
		return smi.BinPath
	}
	return path.Join(hostMountPrefix, smi.BinPath)
}

func (smi *NpuSMI) command(args ...string) ([]byte, error) {
	hostMountPrefix := os.Getenv("HOST_MOUNT_PREFIX")
	cmds := append([]string{smi.BinPath, "info"}, args...)
	if len(hostMountPrefix) > 0 {
		cmds = append([]string{"chroot", hostMountPrefix}, cmds...)
	}
	cmd := execCommand(cmds[0], cmds[1:]...)
	out, err := internal.CombinedOutputTimeout(cmd, time.Duration(smi.Timeout))
	if err != nil {
		return nil, fmt.Errorf("failed to run command %q: %w - %s", strings.Join(cmd.Args, " "), err, string(out))
	}
	return out, nil
}

func (smi *NpuSMI) pollMetrics() ([]*NpuInfo, error) {
	out, err := smi.command("-l")
	if err != nil {
		return nil, err
	}

	npuIDs, err := parseNpuIDs(out)
	if err != nil {
		return nil, err
	}

	results := make([]*NpuInfo, 0, len(npuIDs))
	for _, npuID := range npuIDs {
		info := &NpuInfo{
			NpuID:   npuID,
			Devices: make(map[int]*Device),
		}
		if err := smi.pollNpu(info); err != nil {
			return nil, err
		}
		results = append(results, info)
	}
	return results, nil
}

func (smi *NpuSMI) pollNpu(info *NpuInfo) error {
	npuID := strconv.Itoa(info.NpuID)
	commands := []struct {
		metricType string
		parse      func(*NpuInfo, []byte) error
	}{
		{metricType: "usages", parse: parseUsages},
		{metricType: "board", parse: parseBoard},
		{metricType: "temp", parse: parseTemperature},
		{metricType: "power", parse: parsePower},
	}

	for _, command := range commands {
		out, err := smi.command("-t", command.metricType, "-i", npuID)
		if err != nil {
			return err
		}
		if err := command.parse(info, out); err != nil {
			return fmt.Errorf("parsing %s for NPU %d failed: %w", command.metricType, info.NpuID, err)
		}
	}
	return nil
}

func collectResults(acc telegraf.Accumulator, infos []*NpuInfo) {
	for _, info := range infos {
		for _, dev := range info.Devices {
			acc.AddFields(measurement, dev.getFields(), dev.getTags(info))
		}
	}
}

type NpuInfo struct {
	NpuID           int
	ProductName     string
	Model           string
	Manufacturer    string
	SerialNumber    string
	SoftwareVersion string
	FirmwareVersion string
	PCIeBusInfo     string
	SlotID          string
	Devices         map[int]*Device
}

type Device struct {
	NpuID                 int
	ChipID                int
	DDRCapacity           int64
	DDRUsageRate          float64
	DDRHugepagesTotal     int64
	DDRHugepagesUsageRate float64
	HBMCapacity           int64
	HBMUsageRate          float64
	AICoreUsageRate       float64
	AIVectorUsageRate     float64
	AICpuUsageRate        float64
	CtrlCpuUsageRate      float64
	DDRBandwidthUsageRate float64
	HBMBandwidthUsageRate float64
	NpuUtilization        float64
	AICubeUsageRate       float64
	Temperature           float64
	Power                 float64
}

func (info *NpuInfo) device(chipID int) *Device {
	if dev, found := info.Devices[chipID]; found {
		return dev
	}
	dev := &Device{
		NpuID:  info.NpuID,
		ChipID: chipID,
	}
	info.Devices[chipID] = dev
	return dev
}

func (d *Device) getFields() map[string]interface{} {
	return map[string]interface{}{
		"ddr_capacity":             d.DDRCapacity,
		"ddr_usage_rate":           d.DDRUsageRate,
		"ddr_hugepages_total":      d.DDRHugepagesTotal,
		"ddr_hugepages_usage_rate": d.DDRHugepagesUsageRate,
		"hbm_capacity":             d.HBMCapacity,
		"hbm_usage_rate":           d.HBMUsageRate,
		"aicore_usage_rate":        d.AICoreUsageRate,
		"aivector_usage_rate":      d.AIVectorUsageRate,
		"aicpu_usage_rate":         d.AICpuUsageRate,
		"ctrlcpu_usage_rate":       d.CtrlCpuUsageRate,
		"ddr_bandwidth_usage_rate": d.DDRBandwidthUsageRate,
		"hbm_bandwidth_usage_rate": d.HBMBandwidthUsageRate,
		"npu_utilization":          d.NpuUtilization,
		"aicube_usage_rate":        d.AICubeUsageRate,
		"temperature":              d.Temperature,
		"npu_real_time_power":      d.Power,
	}
}

func (d *Device) getTags(info *NpuInfo) map[string]string {
	tags := map[string]string{
		"npu_id":  strconv.Itoa(d.NpuID),
		"chip_id": strconv.Itoa(d.ChipID),
	}
	addTag(tags, "product_name", info.ProductName)
	addTag(tags, "model", info.Model)
	addTag(tags, "manufacturer", info.Manufacturer)
	addTag(tags, "serial_number", info.SerialNumber)
	addTag(tags, "software_version", info.SoftwareVersion)
	addTag(tags, "firmware_version", info.FirmwareVersion)
	addTag(tags, "pcie_bus_info", info.PCIeBusInfo)
	addTag(tags, "slot_id", info.SlotID)
	return tags
}

func addTag(tags map[string]string, key, value string) {
	if value != "" {
		tags[key] = value
	}
}

func parseNpuIDs(content []byte) ([]int, error) {
	var ids []int
	seen := make(map[int]bool)
	for _, kv := range parseKeyValues(content) {
		if kv.key != "NPU ID" {
			continue
		}
		id, err := strconv.Atoi(kv.value)
		if err != nil {
			return nil, fmt.Errorf("invalid NPU ID %q: %w", kv.value, err)
		}
		if !seen[id] {
			ids = append(ids, id)
			seen[id] = true
		}
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("no NPU IDs found")
	}
	return ids, nil
}

func parseBoard(info *NpuInfo, content []byte) error {
	for _, kv := range parseKeyValues(content) {
		switch kv.key {
		case "NPU ID":
			npuID, err := strconv.Atoi(kv.value)
			if err != nil {
				return fmt.Errorf("invalid NPU ID %q: %w", kv.value, err)
			}
			info.NpuID = npuID
		case "Product Name":
			info.ProductName = kv.value
		case "Model":
			info.Model = kv.value
		case "Manufacturer":
			info.Manufacturer = kv.value
		case "Serial Number":
			info.SerialNumber = kv.value
		case "Software Version":
			info.SoftwareVersion = kv.value
		case "Firmware Version":
			info.FirmwareVersion = kv.value
		case "PCIe Bus Info":
			info.PCIeBusInfo = kv.value
		case "Slot ID":
			info.SlotID = kv.value
		}
	}
	return nil
}

func parseUsages(info *NpuInfo, content []byte) error {
	return parseChipBlocks(content, func(chipID int, values map[string]string) error {
		dev := info.device(chipID)
		dev.NpuID = info.NpuID
		if err := setInt64(values, "DDR Capacity(MB)", &dev.DDRCapacity); err != nil {
			return err
		}
		if err := setFloat(values, "DDR Usage Rate(%)", &dev.DDRUsageRate); err != nil {
			return err
		}
		if err := setInt64(values, "DDR Hugepages Total(page)", &dev.DDRHugepagesTotal); err != nil {
			return err
		}
		if err := setFloat(values, "DDR Hugepages Usage Rate(%)", &dev.DDRHugepagesUsageRate); err != nil {
			return err
		}
		if err := setInt64(values, "HBM Capacity(MB)", &dev.HBMCapacity); err != nil {
			return err
		}
		if err := setFloat(values, "HBM Usage Rate(%)", &dev.HBMUsageRate); err != nil {
			return err
		}
		if err := setFloat(values, "Aicore Usage Rate(%)", &dev.AICoreUsageRate); err != nil {
			return err
		}
		if err := setFloat(values, "Aivector Usage Rate(%)", &dev.AIVectorUsageRate); err != nil {
			return err
		}
		if err := setFloat(values, "Aicpu Usage Rate(%)", &dev.AICpuUsageRate); err != nil {
			return err
		}
		if err := setFloat(values, "Ctrlcpu Usage Rate(%)", &dev.CtrlCpuUsageRate); err != nil {
			return err
		}
		if err := setFloat(values, "DDR Bandwidth Usage Rate(%)", &dev.DDRBandwidthUsageRate); err != nil {
			return err
		}
		if err := setFloat(values, "HBM Bandwidth Usage Rate(%)", &dev.HBMBandwidthUsageRate); err != nil {
			return err
		}
		if err := setFloat(values, "NPU Utilization(%)", &dev.NpuUtilization); err != nil {
			return err
		}
		return setFloat(values, "Aicube Usage Rate(%)", &dev.AICubeUsageRate)
	})
}

func parseTemperature(info *NpuInfo, content []byte) error {
	return parseChipBlocks(content, func(chipID int, values map[string]string) error {
		dev := info.device(chipID)
		return setFloat(values, "NPU Temperature (C)", &dev.Temperature)
	})
}

func parsePower(info *NpuInfo, content []byte) error {
	return parseChipBlocks(content, func(chipID int, values map[string]string) error {
		dev := info.device(chipID)
		return setFloat(values, "NPU Real-time Power(W)", &dev.Power)
	})
}

type keyValue struct {
	key   string
	value string
}

func parseKeyValues(content []byte) []keyValue {
	lines := strings.Split(string(content), "\n")
	values := make([]keyValue, 0, len(lines))
	for _, line := range lines {
		key, value, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			continue
		}
		values = append(values, keyValue{key: key, value: value})
	}
	return values
}

func parseChipBlocks(content []byte, apply func(chipID int, values map[string]string) error) error {
	values := make(map[string]string)
	for _, kv := range parseKeyValues(content) {
		if kv.key == "NPU ID" || kv.key == "Chip Count" {
			continue
		}
		if kv.key != "Chip ID" {
			values[kv.key] = kv.value
			continue
		}

		chipID, err := strconv.Atoi(kv.value)
		if err != nil {
			return fmt.Errorf("invalid Chip ID %q: %w", kv.value, err)
		}
		if err := apply(chipID, values); err != nil {
			return err
		}
		values = make(map[string]string)
	}
	return nil
}

func setFloat(values map[string]string, key string, target *float64) error {
	value, found := values[key]
	if !found || value == "" || value == "N/A" || value == "NA" {
		return nil
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("invalid %s %q: %w", key, value, err)
	}
	*target = parsed
	return nil
}

func setInt64(values map[string]string, key string, target *int64) error {
	value, found := values[key]
	if !found || value == "" || value == "N/A" || value == "NA" {
		return nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid %s %q: %w", key, value, err)
	}
	*target = parsed
	return nil
}
