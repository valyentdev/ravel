package config

type VCpusMemory struct {
	VCpus         int   `json:"vcpus"`
	MemoryConfigs []int `json:"memory_configs"`
}
type MachineResourcesTemplates struct {
	FrequencyByCpu int           `json:"frequency_by_cpu"` // Frequency allocated in MHz
	Combinations   []VCpusMemory `json:"combinations"`
}
