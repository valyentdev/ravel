package machine

import (
	"context"
	"log/slog"

	"github.com/valyentdev/ravel/core/instance"
	"github.com/valyentdev/ravel/runtime"
)

func (m *Machine) prepare(ctx context.Context) error {
	errMsg := "Internal error"

	err := m.state.PushPrepareEvent(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			m.onPrepareFailed(errMsg)
		}
	}()

	_, err = m.runtime.PullImage(ctx, runtime.PullImageOptions{
		Ref: m.state.MachineInstance().Version.Config.Image,
	})
	if err != nil {
		errMsg = "Failed to pull image"
		return err
	}

	mi := m.state.MachineInstance()

	i, err := m.runtime.CreateInstance(ctx, mi.Machine.InstanceId, runtime.InstanceOptions{
		Metadata: instance.InstanceMetadata{
			MachineId:      mi.Machine.Id,
			MachineVersion: mi.Version.Id.String(),
		},
		Config: instance.InstanceConfig{
			Image: mi.Version.Config.Image,
			Guest: instance.InstanceGuestConfig{
				MemoryMB: mi.Version.Resources.MemoryMB,
				Cpus:     mi.Version.Resources.CpuMHz,
				VCpus:    mi.Version.Config.Guest.Cpus,
			},
			Init:       mi.Version.Config.Workload.Init,
			Env:        mi.Version.Config.Workload.Env,
			StopConfig: mi.Version.Config.StopConfig,
		},
	})
	if err != nil {
		errMsg = "Failed to create instance"
		return err
	}

	err = m.state.PushPreparedEvent(ctx, i.Network)
	if err != nil {
		slog.Error("Failed to push prepared event", "error", err)
		return err
	}

	return nil
}

func (m *Machine) onPrepareFailed(msg string) {
	if err := m.state.PushPrepareFailedEvent(context.Background(), msg); err != nil {
		slog.Error("Failed to push PrepareFailed event", "error", err)
	}

	m.destroyImpl(context.Background())

}
