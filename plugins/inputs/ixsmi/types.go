package ixsmi

type ixsmiLog struct {
	DriverVersion string `xml:"driver_version"`
	CUDAVersion   string `xml:"cuda_version"`
	GPU           []gpu  `xml:"gpu"`
}

type gpu struct {
	ID           string           `xml:"id,attr"`
	ProdName     string           `xml:"product_name"`
	Serial       string           `xml:"serial"`
	UUID         string           `xml:"uuid"`
	MinorNumber  string           `xml:"minor_number"`
	VbiosVersion string           `xml:"vbios_version"`
	FanSpeed     string           `xml:"fan_speed"`
	ComputeMode  string           `xml:"compute_mode"`
	PState       string           `xml:"performance_state"`
	PCI          pciInfo          `xml:"pci"`
	Memory       memoryStats      `xml:"memory_usage"`
	Utilization  utilizationStats `xml:"utilization"`
	EccMode      eccMode          `xml:"ecc_mode"`
	EccErrors    eccErrors        `xml:"ecc_errors"`
	Temp         tempStats        `xml:"temperature"`
	Power        powerReadings    `xml:"power_readings"`
	Clocks       clockStats       `xml:"clocks"`
}

type pciInfo struct {
	LinkInfo struct {
		PCIEGen struct {
			CurrentLinkGen string `xml:"current_link_gen"`
		} `xml:"pcie_gen"`
		LinkWidth struct {
			CurrentLinkWidth string `xml:"current_link_width"`
		} `xml:"link_widths"`
	} `xml:"pci_gpu_link_info"`
}

type memoryStats struct {
	Total string `xml:"total"`
	Used  string `xml:"used"`
	Free  string `xml:"free"`
}

type utilizationStats struct {
	GPU     string `xml:"gpu_util"`
	Memory  string `xml:"memory_util"`
	Encoder string `xml:"encoder_util"`
	Decoder string `xml:"decoder_util"`
}

type eccMode struct {
	CurrentEcc string `xml:"current_ecc"`
}

type eccErrors struct {
	SingleBit string `xml:"single_bit"`
	DoubleBit string `xml:"double_bit"`
}

type tempStats struct {
	GPUTemp    string `xml:"gpu_temp"`
	MemoryTemp string `xml:"memory_temp"`
}

type powerReadings struct {
	GPUPowerDraw           string `xml:"gpu_power_draw"`
	CurrentGPUPowerLimit   string `xml:"current_gpu_power_limit"`
	BoardPowerDraw         string `xml:"board_power_draw"`
	CurrentBoardPowerLimit string `xml:"current_board_power_limit"`
}

type clockStats struct {
	SM     string `xml:"sm_clock"`
	Memory string `xml:"mem_clock"`
}
