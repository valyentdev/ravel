package container

import (
	"context"
	"fmt"
)

func (r *Runtime) SubscribeToLogs(ctx context.Context, machineId string, ch chan []byte) error {
	rm, ok := r.runningVMs[machineId]
	if !ok {
		return fmt.Errorf("machine %q is not running", machineId)
	}

	rm.SubscribeToLogs(ctx, ch)

	return nil
}

func (r *Runtime) GetLogs(machineId string) ([]byte, error) {
	rm, ok := r.runningVMs[machineId]
	if !ok {
		return nil, fmt.Errorf("machine %q is not running", machineId)
	}

	return rm.GetLog(), nil
}
