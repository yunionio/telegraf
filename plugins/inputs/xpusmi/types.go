package xpusmi

type xpusmiLog struct {
	DriverVersion string `xml:"driver_version"`
	CUDAVersion   string `xml:"cuda_version"`
	Xpu           []xpu  `xml:"xpu"`
}

type xpu struct {
	ID                  string           `xml:"id,attr"`
	ProductName         string           `xml:"product_name"`
	ProductBrand        string           `xml:"product_brand"`
	ProductArchitecture string           `xml:"product_architecture"`
	Serial              string           `xml:"serial"`
	UUID                string           `xml:"uuid"`
	MinorNumber         string           `xml:"minor_number"`
	OamID               string           `xml:"oam_id"`
	PCI                 pciInfo          `xml:"pci"`
	Memory              memoryStats      `xml:"fb_memory_usage"`
	L3                  memoryStats      `xml:"l3_usage"`
	Utilization         utilizationStats `xml:"utilization"`
	EccMode             eccMode          `xml:"ecc_mode"`
	EccErrors           eccErrors        `xml:"ecc_errors"`
	Temp                tempStats        `xml:"temperature"`
	Power               powerReadings    `xml:"power_readings"`
	Clocks              clockStats       `xml:"clocks"`
}

type pciInfo struct {
	LinkInfo struct {
		PCIEGen struct {
			CurrentLinkGen string `xml:"current_link_gen"`
		} `xml:"pcie_gen"`
		LinkWidth struct {
			CurrentLinkWidth string `xml:"current_link_width"`
		} `xml:"link_widths"`
	} `xml:"pci_xpu_link_info"`
}

type memoryStats struct {
	Total string `xml:"total"`
	Used  string `xml:"used"`
	Free  string `xml:"free"`
}

type utilizationStats struct {
	XpuUtil string `xml:"xpu_util"`
}

type eccMode struct {
	CurrentEcc string `xml:"current_ecc"`
}

type eccErrors struct {
	Volatile struct {
		DramCorrectable   string `xml:"dram_correctable"`
		DramUncorrectable string `xml:"dram_uncorrectable"`
	} `xml:"volatile"`
	Aggregate struct {
		DramCorrectable   string `xml:"dram_correctable"`
		DramUncorrectable string `xml:"dram_uncorrectable"`
	} `xml:"aggregate"`
}

type tempStats struct {
	XpuTemp string `xml:"xpu_temp"`
}

type powerReadings struct {
	EnforcedPowerLimit string `xml:"enforced_power_limit"`
	PowerDraw          string `xml:"power_draw"`
}

type clockStats struct {
	ClusterClock string `xml:"cluster_clock"`
	CdnnClock    string `xml:"cdnn_clock"`
}
