package instancerunner

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/api/errdefs"
	"github.com/valyentdev/ravel/core/instance"
	"github.com/valyentdev/ravel/runtime/drivers"
	"github.com/valyentdev/ravel/runtime/logging"
)

type vmRunner struct {
	driver     drivers.Driver
	logger     *logging.InstanceLogger
	i          instance.Instance
	hasStarted atomic.Bool
	vm         drivers.InstanceTask
	waitCh     chan struct{}
	exitResult instance.ExitResult
}

func (r *vmRunner) terminated() bool {
	select {
	case <-r.waitCh:
		return true
	default:
		return false
	}
}

func newVMRunner(
	i instance.Instance,
	logger *logging.InstanceLogger,
	driver drivers.Driver,
) *vmRunner {
	return &vmRunner{
		i:      i,
		driver: driver,
		logger: logger,
		waitCh: make(chan struct{}),
	}
}

func (r *vmRunner) Recover() error {
	vm, err := r.driver.RecoverInstanceTask(context.Background(), &r.i)
	if err != nil {
		slog.Error("failed to recover vm", "error", err)
		cerr := r.driver.CleanupInstanceTask(context.Background(), &r.i)
		if cerr != nil {
			slog.Error("failed to cleanup vm", "error", cerr)
		}
		return err
	}

	r.vm = vm
	go r.run()
	r.hasStarted.Store(true)

	return nil
}

func (r *vmRunner) Stop(signal string, timeout time.Duration) error {
	if r.terminated() {
		return nil
	}
	ctx := context.Background()
	err := r.vm.Stop(context.Background(), signal)
	if err != nil {
		slog.Error("failed to stop vm", "error", err)
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	exited := r.vm.WaitExit(ctxTimeout)
	if exited {
		<-r.waitCh
		return nil
	}

	slog.Debug("vm did not exit in time, shutting down the vmm")

	err = r.vm.Shutdown(ctx)
	if err != nil {
		return err
	}

	<-r.waitCh

	return nil
}

func (r *vmRunner) Start() error {
	ctx := context.Background()
	slog.Debug("ensuring instance network")
	vm, err := r.driver.BuildInstanceTask(ctx, &r.i)
	if err != nil {
		slog.Error("failed to build vm", "error", err)
		return err
	}
	defer func() {
		if err != nil {
			err := r.driver.CleanupInstanceTask(ctx, &r.i)
			if err != nil {
				slog.Error("failed to cleanup vm", "error", err)
			}
		}
	}()

	r.vm = vm

	slog.Debug("starting vm")
	err = vm.Start(ctx)
	if err != nil {
		return err
	}

	r.hasStarted.Store(true)

	go r.run()
	return nil
}

func getLogFile(id string) string {
	return fmt.Sprintf("/var/lib/ravel/instances/%s/vm.logs", id)
}

func (r *vmRunner) run() {
	err := r.logger.Start(getLogFile(r.i.Id))
	if err != nil {
		slog.Error("failed to start logger", "error", err)
		err = nil // ignore we must continue
	}

	defer r.logger.Stop()

	result := r.vm.Run()
	r.exitResult = result

	slog.Debug("vm exited", "exitCode", result.ExitCode, "instance", r.i.Id)

	err = r.driver.CleanupInstanceTask(context.Background(), &r.i)
	if err != nil {
		slog.Error("failed to cleanup vm", "error", err)
	}

	close(r.waitCh)
}

func (r *vmRunner) Run() instance.ExitResult {
	<-r.waitCh
	return r.exitResult
}

func (r *vmRunner) Wait() <-chan struct{} {
	return r.waitCh
}

func (r *vmRunner) Exec(ctx context.Context, cmd []string, timeout time.Duration) (*api.ExecResult, error) {
	if !r.hasStarted.Load() || r.terminated() {
		return nil, errdefs.NewFailedPrecondition("instance is not running")
	}
	return r.vm.Exec(ctx, cmd, timeout)
}

func (r *vmRunner) Signal(ctx context.Context, signal string) error {
	if !r.hasStarted.Load() || r.terminated() {
		return errdefs.NewFailedPrecondition("instance is not running")
	}
	return r.vm.Signal(ctx, signal)
}
