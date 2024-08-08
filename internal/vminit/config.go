package vminit

type (
	IPConfig struct {
		IPNet     string
		Broadcast string
		Gateway   string
	}

	NetworkConfig struct {
		IPConfigs      []IPConfig
		DefaultGateway string
	}

	ImageConfig struct {
		Cmd        []string
		Entrypoint []string
		Env        []string
		WorkingDir *string
		User       *string
	}

	Mounts struct {
		MountPath  string
		DevicePath string
	}

	EtcResolv struct {
		Nameservers []string
	}

	EtcHost struct {
		Host string
		IP   string
		Desc string
	}

	Config struct {
		ImageConfig        *ImageConfig
		UserOverride       string
		EntrypointOverride []string
		CmdOverride        []string
		Hostname           string
		ExtraEnv           []string
		RootDevice         string
		Mounts             []Mounts

		EtcResolv EtcResolv
		EtcHost   []EtcHost
		Network   NetworkConfig
	}
)
