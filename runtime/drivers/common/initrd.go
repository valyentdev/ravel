package common

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/alexisbouchez/ravel/core/instance"
	"github.com/alexisbouchez/ravel/initd"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/u-root/u-root/pkg/cpio"
	"golang.org/x/sys/unix"
)

// StringPtr returns a pointer to the given string.
func StringPtr(s string) *string {
	return &s
}

// WriteInitrd writes an initramfs to the given file.
func WriteInitrd(file *os.File, initBinaryPath string, inst *instance.Instance, image v1.Image) error {
	slog.Debug("writing initrd", "instance", inst.Id)

	t1 := time.Now()
	initFile, err := os.Open(initBinaryPath)
	if err != nil {
		return fmt.Errorf("failed to read init binary: %w", err)
	}
	defer initFile.Close()

	slog.Info("init binary opened after", "time", time.Since(t1))

	gz := gzip.NewWriter(file)
	defer gz.Close()

	w := cpio.Newc.Writer(gz)

	config := GetInitConfig(inst, image.Config)
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal init config: %w", err)
	}

	initInfos, err := initFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get init stat: %w", err)
	}

	configRecord := cpio.StaticFile("/ravel/run.json", string(configJSON), 0644)

	initRecord := cpio.Record{
		ReaderAt: initFile,
		Info: cpio.Info{
			FileSize: uint64(initInfos.Size()),
			Name:     "ravel-init",
			Mode:     unix.S_IFREG | 0755,
		},
	}

	slog.Info("init and config records created after", "time", time.Since(t1))
	err = cpio.WriteRecordsAndDirs(w, []cpio.Record{initRecord, configRecord})
	if err != nil {
		return fmt.Errorf("failed to write records and dirs: %w", err)
	}

	slog.Info("initrd written after", "time", time.Since(t1))

	return nil
}

// GetInitConfig creates an initd configuration from an instance and image config.
func GetInitConfig(inst *instance.Instance, image v1.ImageConfig) initd.Config {
	config := inst.Config

	mounts := GetAdditionalMounts(config.Mounts)

	return initd.Config{
		ImageConfig: &initd.ImageConfig{
			User:       StringPtr(image.User),
			WorkingDir: StringPtr(image.WorkingDir),
			Cmd:        image.Cmd,
			Entrypoint: image.Entrypoint,
			Env:        image.Env,
		},
		UserOverride:       config.Init.User,
		CmdOverride:        config.Init.Cmd,
		EntrypointOverride: config.Init.Entrypoint,
		RootDevice:         "/dev/vda",
		EtcResolv: initd.EtcResolv{
			Nameservers: []string{"8.8.8.8"},
		},
		ExtraEnv: config.Env,
		Network: initd.NetworkConfig{
			IPConfigs: []initd.IPConfig{
				{
					IPNet:     inst.Network.Local.InstanceIPNet().String(),
					Broadcast: inst.Network.Local.Broadcast.String(),
					Gateway:   inst.Network.Local.Gateway.String(),
				},
			},
			DefaultGateway: inst.Network.Local.Gateway.String(),
		},
		Mounts: mounts,
	}
}
