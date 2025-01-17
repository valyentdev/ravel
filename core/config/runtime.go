package config

type RuntimeConfig struct {
	CloudHypervisorBinary string `json:"cloud_hypervisor_binary" toml:"cloud_hypervisor_binary"`
	JailerBinary          string `json:"jailer_binary" toml:"jailer_binary"`
	InitBinary            string `json:"init_binary" toml:"init_binary"`
	LinuxKernel           string `json:"linux_kernel" toml:"linux_kernel"`
	ZFSPool               string `json:"zfs_pool" toml:"zfs_pool"`
}
