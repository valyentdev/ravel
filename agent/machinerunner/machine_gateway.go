package machinerunner

func (m *MachineRunner) EnableGateway() error {
	if m.state.State().MachineGatewayEnabled {
		return nil
	}

	if err := m.state.EnableGateway(); err != nil {
		return err
	}

	return nil
}

func (m *MachineRunner) DisableGateway() error {
	if !m.state.State().MachineGatewayEnabled {
		return nil
	}

	if err := m.state.DisableGateway(); err != nil {
		return err
	}

	return nil
}
