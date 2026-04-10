package config

// RuntimeType specifies which VM runtime to use.
type RuntimeType string

const (
	RuntimeTypeCloudHypervisor RuntimeType = "cloudhypervisor"
	RuntimeTypeFirecracker     RuntimeType = "firecracker"
)

type RuntimeConfig struct {
	// RuntimeType specifies which VM runtime to use: "cloudhypervisor" (default) or "firecracker"
	RuntimeType           RuntimeType `json:"runtime_type" toml:"runtime_type"`
	CloudHypervisorBinary string      `json:"cloud_hypervisor_binary" toml:"cloud_hypervisor_binary"`
	FirecrackerBinary     string      `json:"firecracker_binary" toml:"firecracker_binary"`
	JailerBinary          string      `json:"jailer_binary" toml:"jailer_binary"`
	InitBinary            string      `json:"init_binary" toml:"init_binary"`
	LinuxKernel           string      `json:"linux_kernel" toml:"linux_kernel"`
	ZFSPool               string      `json:"zfs_pool" toml:"zfs_pool"`
}

// GetRuntimeType returns the runtime type, defaulting to CloudHypervisor if not specified.
func (c *RuntimeConfig) GetRuntimeType() RuntimeType {
	if c.RuntimeType == "" {
		return RuntimeTypeCloudHypervisor
	}
	return c.RuntimeType
}
