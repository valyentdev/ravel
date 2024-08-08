package instance

import "context"

func (m *Manager) Wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-m.waitCh:
	}
	return nil
}
