package config

type RuntimeConfig struct {
	InitBinary  string `json:"init_binary" toml:"init_binary"`
	LinuxKernel string `json:"linux_kernel" toml:"linux_kernel"`
	Snapshotter string `json:"snapshotter" toml:"snapshotter"`
}
