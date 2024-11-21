package instancemanager

import (
	"context"
	"fmt"

	"github.com/valyentdev/ravel/core/errdefs"
	"github.com/valyentdev/ravel/core/instance"
)

func (m *Manager) Destroy(ctx context.Context) error {
	m.state.Lock()
	defer m.state.Unlock()
	i := m.state.Instance()

	status := m.state.Status()

	if status != instance.InstanceStatusStopped {
		return errdefs.NewFailedPrecondition(fmt.Sprintf("instance is in %s state", status))
	}

	err := m.state.UpdateInstanceState(instance.State{
		Status: instance.InstanceStatusDestroying,
	})
	if err != nil {
		return err
	}

	err = m.vmBuilder.CleanupInstance(ctx, &i)
	if err != nil {
		return err
	}

	return nil
}
