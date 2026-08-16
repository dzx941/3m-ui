package system

type CPUInfo struct {
	Percent float64 `json:"percent"`
}

type MemoryInfo struct {
	Used    float64 `json:"used"`    // in MB
	Total   float64 `json:"total"`   // in MB
	Percent float64 `json:"percent"` // in %
}

type DiskInfo struct {
	Used    float64 `json:"used"`    // in GB
	Total   float64 `json:"total"`   // in GB
	Percent float64 `json:"percent"` // in %
}

type NetworkInfo struct {
	Upload   float64 `json:"upload"`   // bytes/sec
	Download float64 `json:"download"` // bytes/sec
}

type SystemStats struct {
	CPU     CPUInfo     `json:"cpu"`
	Memory  MemoryInfo  `json:"memory"`
	Disk    DiskInfo    `json:"disk"`
	Network NetworkInfo `json:"network"`
}
