package container

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/containerd/errdefs"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/u-root/u-root/pkg/cpio"
	"github.com/valyentdev/ravel/internal/vminit"
	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/helper/cloudhypervisor"
	"golang.org/x/sys/unix"
)

func (r *Runtime) PrepareInstance(ctx context.Context, instance core.Instance) (err error, fatal bool) {
	config := instance.Config

	var auth string

	slog.Debug("Pulling image", "image", config.Workload.Image)
	image, err := r.images.Pull(ctx, config.Workload.Image, &core.RegistryAuthConfig{
		Auth: auth,
	})
	if err != nil {
		if errdefs.IsNotFound(err) || errdefs.IsFailedPrecondition(err) {
			fatal = true
		}
		return
	}

	v1Image, err := image.Spec(ctx)
	if err != nil {
		return
	}

	err = r.prepareInitRD(instance, v1Image.Config)
	if err != nil {
		fatal = true
		return
	}

	slog.Debug("Pulled image", "image", image.Name())

	return
}

func (r Runtime) prepareInitRD(instance core.Instance, image v1.ImageConfig) error {
	init, err := os.Open(r.config.InitBinary)
	if err != nil {
		return fmt.Errorf("failed to read init binary: %w", err)
	}
	defer init.Close()

	initrdPath := getInitrdPath(instance.Id)
	err = os.MkdirAll(getInstancePath(instance.Id), 0755)
	if err != nil {
		return fmt.Errorf("failed to create instance directory: %w", err)
	}

	file, err := os.Create(initrdPath)
	if err != nil {
		return fmt.Errorf("failed to create initrd file: %w", err)
	}

	gz := gzip.NewWriter(file)
	defer gz.Close()

	w := cpio.Newc.Writer(gz)

	config := getInitConfig(instance, image)
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal init config: %w", err)
	}

	initInfos, err := init.Stat()
	if err != nil {
		return fmt.Errorf("failed to get init stat: %w", err)
	}

	configRecord := cpio.StaticFile("/ravel/run.json", string(configJSON), 0644)

	initRecord := cpio.Record{
		ReaderAt: init,
		Info: cpio.Info{
			FileSize: uint64(initInfos.Size()),
			Name:     "ravel-init",
			Mode:     unix.S_IFREG | 0755,
		},
	}

	err = cpio.WriteRecordsAndDirs(w, []cpio.Record{initRecord, configRecord})
	if err != nil {
		return fmt.Errorf("failed to write records and dirs: %w", err)
	}

	return nil
}

func getInitConfig(instance core.Instance, image v1.ImageConfig) vminit.Config {
	config := instance.Config
	return vminit.Config{
		ImageConfig: &vminit.ImageConfig{
			User:       cloudhypervisor.StringPtr(image.User),
			WorkingDir: cloudhypervisor.StringPtr(image.WorkingDir),
			Cmd:        image.Cmd,
			Entrypoint: image.Entrypoint,
			Env:        image.Env,
		},
		UserOverride:       config.Workload.Init.User,
		CmdOverride:        config.Workload.Init.Cmd,
		EntrypointOverride: config.Workload.Init.Entrypoint,
		RootDevice:         "/dev/vda",
		EtcResolv: vminit.EtcResolv{
			Nameservers: []string{"8.8.8.8"},
		},
		ExtraEnv: config.Workload.Env,
		Network:  vminit.NetworkConfig{},
	}
}
