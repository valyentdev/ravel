package container

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

func (r *Runtime) StopVM(ctx context.Context, instanceId string, signal string, timeout time.Duration) error {
	rm, ok := r.runningVMs[instanceId]
	if !ok {
		return fmt.Errorf("machine %q is not running", instanceId)
	}

	var signalString string

	slog.Info("stopping instance", "instance", instanceId, "signal", signalString)
	err := rm.Stop(ctx, signal)
	if err != nil {
		return err
	}

	timeoutCTX, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	exited := rm.WaitExit(timeoutCTX)
	if !exited {
		err := rm.Shutdown(context.Background())
		if err != nil {
			slog.Error("failed to shutdown machine", "machine", instanceId, "instance", instanceId, "err", err)
		}
	}

	return nil
}
