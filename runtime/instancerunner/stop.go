package instancerunner

import (
	"context"
	"fmt"
	"time"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/api/errdefs"
	"github.com/valyentdev/ravel/core/instance"
)

func mergeStopConfigs(defaultConfig *api.StopConfig, configs ...*api.StopConfig) *api.StopConfig {
	if defaultConfig == nil {
		panic("defaultConfig is nil")
	}

	config := &api.StopConfig{
		Signal:  defaultConfig.Signal,
		Timeout: defaultConfig.Timeout,
	}

	for _, c := range configs {
		if c == nil {
			continue
		}
		if c.Signal != nil {
			config.Signal = c.Signal
		}

		if c.Timeout != nil {
			config.Timeout = c.Timeout
		}
	}

	return config
}

func (ir *InstanceRunner) getStopParams(config *api.StopConfig) (string, time.Duration) {
	defaultStopConfig := api.GetDefaultStopConfig()
	storedConfig := ir.Instance().Config.Stop

	config = mergeStopConfigs(defaultStopConfig, storedConfig, config)

	var signal string
	var timeout time.Duration

	signal = *config.Signal
	timeout = time.Duration(*config.Timeout) * time.Second

	return signal, timeout
}

func (ir *InstanceRunner) Stop(ctx context.Context, config *api.StopConfig) error {
	var willStop bool
	ir.lock()
	defer func() {
		if !willStop {
			ir.unlock()
		}
	}()

	status := ir.Status()

	if status == instance.InstanceStatusStopped {
		return nil
	}

	if status != instance.InstanceStatusRunning {
		return errdefs.NewFailedPrecondition(fmt.Sprintf("instance is in %s status", status))
	}

	if err := ir.updateInstanceStateFunc(func(state *instance.State) {
		state.Stopping = true
	}); err != nil {
		return err
	}
	// We need to unlock the state before calling stopImpl to allow user to send another stop request with different signal or timeout
	ir.unlock()
	willStop = true

	signal, timeout := ir.getStopParams(config)

	return ir.stopImpl(signal, timeout)
}

func (ir *InstanceRunner) stopImpl(signal string, timeout time.Duration) error {
	runner := ir.getVMRunner()
	if runner == nil {
		return nil
	}

	err := ir.runner.Stop(signal, timeout)
	if err != nil {
		return err
	}
	return nil
}
