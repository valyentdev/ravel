package machinerunner

import "context"

func (m *MachineRunner) EnableGateway(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.state.MachineInstance().State.MachineGatewayEnabled {
		return nil
	}

	if err := m.state.PushGatewayEvent(true); err != nil {
		return err
	}

	return nil
}

func (m *MachineRunner) DisableGateway(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.state.MachineInstance().State.MachineGatewayEnabled {
		return nil
	}

	if err := m.state.PushGatewayEvent(false); err != nil {
		return err
	}

	return nil
}
