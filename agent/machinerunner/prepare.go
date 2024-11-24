package machinerunner

import (
	"context"
	"log/slog"

	"github.com/valyentdev/ravel/core/daemon"
)

func (m *MachineRunner) prepare(ctx context.Context) error {
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

	_, err = m.runtime.PullImage(ctx, daemon.ImagePullOptions{
		Ref: m.state.MachineInstance().Version.Config.Image,
	})
	if err != nil {
		errMsg = "Failed to pull image"
		return err
	}

	mi := m.state.MachineInstance()

	i, err := m.runtime.CreateInstance(ctx, mi.InstanceOptions())
	if err != nil {
		errMsg = "Failed to create instance"
		return err
	}

	err = m.state.PushPreparedEvent(ctx, i.Network)
	if err != nil {
		slog.Error("Failed to push prepared event", "error", err)
	}

	return nil
}

func (m *MachineRunner) onPrepareFailed(msg string) {
	if err := m.state.PushPrepareFailedEvent(context.Background(), msg); err != nil {
		slog.Error("Failed to push PrepareFailed event", "error", err)
	}

	m.destroyImpl(context.Background())
}
