package config

// Server
const (
	defaultServerAddress = ":3000"
)

// Agent
const (
	defaultAgentAddress = "localhost"
	defaultAgentPort    = 8080
	defaultInitBinary   = "/etc/ravel/ravel-init"
	defaultLinuxKernel  = "/etc/ravel/vmlinux.bin"
	defaultAgentDBPath  = "/var/lib/ravel/agent.db"
)

func SetDefaults(config *RavelConfig) {
	if config.Server.Address == "" {
		config.Server.Address = defaultServerAddress
	}

	if config.Agent.Address == "" {
		config.Agent.Address = defaultAgentAddress
	}

	if config.Agent.InitBinary == "" {
		config.Agent.InitBinary = defaultInitBinary
	}

	if config.Agent.LinuxKernel == "" {
		config.Agent.LinuxKernel = defaultLinuxKernel
	}

	if config.Agent.DatabasePath == "" {
		config.Agent.DatabasePath = defaultAgentDBPath
	}
}
