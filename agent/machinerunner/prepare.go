package machinerunner

import (
	"context"
	"log/slog"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/daemon"
)

func (m *MachineRunner) prepare(ctx context.Context) {
	var err error
	errMsg := ""

	defer func() {
		if err != nil {
			m.state.PushPrepareFailedEvent(errMsg)
		}
	}()
	_, err = m.runtime.PullImage(ctx, daemon.ImagePullOptions{
		Ref: m.state.MachineInstance().Version.Config.Image,
	})
	if err != nil {
		errMsg = "Failed to pull image"
		return
	}

	mi := m.state.MachineInstance()

	_, err = m.runtime.CreateInstance(ctx, mi.InstanceOptions())
	if err != nil {
		errMsg = "Failed to create instance"
		return
	}

	m.state.PushPreparedEvent()

}

func (m *MachineRunner) onPrepareFailed(msg string) {
	if _, _, err := m.state.PushPrepareFailedEvent(msg); err != nil {
		slog.Error("Failed to push PrepareFailed event", "error", err)
	}

	m.state.PushDestroyEvent(api.OriginRavel, true, false, "failed to prepare machine")

}
