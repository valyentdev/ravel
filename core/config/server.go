package config

type VCpusMemory struct {
	VCpus         int   `json:"vcpus" toml:"vcpus"`
	MemoryConfigs []int `json:"memory_configs" toml:"memory_configs"`
}
type MachineResourcesTemplates struct {
	VCPUFrequency int           `json:"vcpu_frequency" toml:"vcpu_frequency"`
	Combinations  []VCpusMemory `json:"combinations" toml:"combinations"`
}

type ServerConfig struct {
	Address          string                               `json:"address" toml:"address"`
	PostgresURL      string                               `json:"postgres_url" toml:"postgres_url"`
	MachineTemplates map[string]MachineResourcesTemplates `json:"machine_templates" toml:"machine_templates"`
}
