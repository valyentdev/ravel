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
	PostgresURL        string                               `json:"postgres_url" toml:"postgres_url"`
	MachineTemplates   map[string]MachineResourcesTemplates `json:"machine_templates" toml:"machine_templates"`
	TLS                *TLSConfig                           `json:"tls" toml:"tls"` // used for client auth against agents internal API
	API                ServerAPIConfig                      `json:"api" toml:"api"`
	MainRegistry       string                               `json:"main_registry" toml:"main_registry"`
	NamespacedRegistry bool                                 `json:"namespaced_registry" toml:"namespaced_registry"` // if true, ravel doesnt pull images from main registry if the repository name is different from the namespace
}

type ServerAPIConfig struct {
	Address     string     `json:"address" toml:"address"`
	BearerToken string     `json:"bearer_token" toml:"bearer_token"`
	TLS         *TLSConfig `json:"tls" toml:"tls"` // used to protect the API with (m)TLS
}
