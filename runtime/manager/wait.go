package instancemanager

import "context"

func (m *Manager) waitExit() {
	<-m.waitCh
}

func (m *Manager) WaitExit(ctx context.Context) error {
	select {
	case <-m.waitCh:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
