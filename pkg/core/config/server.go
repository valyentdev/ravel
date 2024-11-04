package config

type VCpusMemory struct {
	VCpus         int   `json:"vcpus" toml:"vcpus"`
	MemoryConfigs []int `json:"memory_configs" toml:"memory_configs"`
}
type MachineResourcesTemplates struct {
	FrequencyByCpu int           `json:"frequency_by_vcpu" toml:"frequency_by_vcpu"` // Frequency allocated in MHz
	Combinations   []VCpusMemory `json:"combinations" toml:"combinations"`
}

type ServerConfig struct {
	Address          string                               `json:"address" toml:"address"`
	PostgresURL      string                               `json:"postgres_url" toml:"postgres_url"`
	MachineTemplates map[string]MachineResourcesTemplates `json:"machine_templates" toml:"machine_templates"`
}
