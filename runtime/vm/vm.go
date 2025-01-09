package vm

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"sync/atomic"
	"time"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/instance"
	"github.com/valyentdev/ravel/initd/client"
	"github.com/valyentdev/ravel/pkg/cloudhypervisor"
)

type vm struct {
	id                     string
	cmd                    *exec.Cmd
	vmm                    *cloudhypervisor.VMM
	runResult              *RunResult
	initClient             *client.InternalClient
	successFullyShutdowned atomic.Bool
	vmConfig               cloudhypervisor.VmConfig
	stopRequested          bool
	waitChan               chan struct{}
}

var _ instance.VM = (*vm)(nil)

func (vm *vm) Id() string {
	return vm.id
}

func newVM(id string, cmd *exec.Cmd, vmConfig cloudhypervisor.VmConfig) *vm {
	vmm := cloudhypervisor.NewVMMClient(getAPISocketPath(id))

	client := client.NewInternalClient(getVsockPath(id))

	return &vm{
		id:         id,
		cmd:        cmd,
		vmConfig:   vmConfig,
		vmm:        vmm,
		waitChan:   make(chan struct{}),
		initClient: client,
	}
}

func (vm *vm) Start(ctx context.Context) error {
	err := vm.cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start vmm for machine %q: %w", vm.Id(), err)
	}
	defer func() {
		if err != nil {
			vm.vmm.ShutdownVMM(ctx)
		}
	}()

	err = vm.vmm.WaitReady(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for vmm to be ready for machine %q: %w", vm.Id(), err)
	}

	err = vm.vmm.CreateVM(ctx, vm.vmConfig)
	if err != nil {
		return fmt.Errorf("failed to create vm for machine %q: %w", vm.Id(), err)
	}

	err = vm.vmm.BootVM(ctx)
	if err != nil {
		return fmt.Errorf("failed to boot vm for machine %q: %w", vm.Id(), err)
	}

	go vm.run()

	return nil

}

func (vm *vm) Signal(ctx context.Context, signal string) error {
	sig := syscallSignal(signal)

	err := vm.initClient.Signal(ctx, int(sig))
	if err != nil {
		return fmt.Errorf("failed to send signal to init: %w", err)
	}

	return nil
}

type RunResult struct {
	HasBeenStopped bool
	VMExited       bool
	InitFailed     bool
	ProcessExited  bool
	ExitCode       int
	ExitedAt       time.Time
}

func (vm *vm) run() {
	slog.Debug("vm run")
	defer close(vm.waitChan)
	defer vm.Shutdown(context.Background())
	result := &RunResult{
		ExitCode: -1,
	}
	defer func() {
		if vm.stopRequested {
			result.HasBeenStopped = true
		}
		vm.runResult = result
	}()

	started := false

	for i := 0; i < 5; i++ {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		err := vm.initClient.HealthCheck(timeoutCtx)
		cancel()
		if err == nil {
			started = true
			break
		}
		slog.Debug("waiting for init to start", "err", err, "id", vm.id)
		time.Sleep(100 * time.Millisecond)
	}

	if !started {
		result.InitFailed = true
		result.ExitedAt = time.Now()
		vm.runResult = result
		return
	}

	slog.Debug("vm has started")

	retryCount := 0

RETRY:
	exitResult, err := vm.initClient.Wait(context.Background())
	if err != nil {
		err := vm.initClient.HealthCheck(context.Background())
		if err != nil {
			result.VMExited = true
			result.ExitedAt = time.Now()
			return
		}
		if retryCount < 10 {
			slog.Debug("retrying to wait for init to exit")
			retryCount++
		} else {
			slog.Debug("failed to wait for init to exit", "err", err)
			result.InitFailed = true
			result.ExitedAt = time.Now()
			return
		}
		goto RETRY
	}

	result.ProcessExited = true
	result.ExitCode = exitResult.ExitCode
	slog.Debug("init process exited", "exitCode", exitResult.ExitCode)
}

func (vm *vm) Shutdown(ctx context.Context) error {
	if vm.successFullyShutdowned.Load() {
		return nil
	}
	err := vm.vmm.ShutdownVMM(ctx)
	if err != nil {
		return fmt.Errorf("failed to shutdown VMM: %w", err)
	}
	vm.successFullyShutdowned.Store(true)
	return nil
}

func (vm *vm) Stop(ctx context.Context, signal string) error {
	vm.stopRequested = true
	err := vm.Signal(ctx, signal)
	if err != nil {
		return fmt.Errorf("failed to send signal to init: %w", err)
	}

	return nil
}

func (vm *vm) Run() instance.ExitResult {
	<-vm.waitChan
	er := instance.ExitResult{
		Success:   vm.runResult.ExitCode == 0,
		ExitCode:  vm.runResult.ExitCode,
		ExitedAt:  vm.runResult.ExitedAt,
		Requested: vm.runResult.HasBeenStopped,
	}
	return er
}

func (vm *vm) WaitExit(ctx context.Context) (exited bool) {
	for {
		select {
		case <-ctx.Done():
			return false
		case <-vm.waitChan:
			return true
		}
	}
}

func (vm *vm) recover() bool {
	ok := false
	state := vm.determinateState()

	if !state.isVMMRunning {
		return ok
	}
	if !state.isVMRunning {
		err := vm.Shutdown(context.Background())
		if err != nil {
			slog.Error("failed to shutdown VMM", "err", err)
		}
		return ok
	}

	go vm.run()

	return true
}

type internalState struct {
	isVMMRunning bool
	isVMRunning  bool
	vminfo       *cloudhypervisor.VmInfo
}

func (vm *vm) determinateState() internalState {
	s := internalState{
		isVMMRunning: false,
		isVMRunning:  false,
	}

	if _, err := vm.vmm.PingVMM(context.Background()); err != nil {
		slog.Debug("failed to ping VMM", "err", err)
		return s
	}
	s.isVMMRunning = true

	vminfo, err := vm.vmm.VMInfo(context.Background())
	if err != nil {
		slog.Debug("failed to get VM info", "err", err)
		return s
	}

	if vminfo.State != cloudhypervisor.Running {
		slog.Debug("VM is not running", "state", vminfo.State)
		return s
	}
	s.isVMRunning = true

	s.vminfo = vminfo

	return s
}

func (vm *vm) Exec(ctx context.Context, cmd []string, timeout time.Duration) (*api.ExecResult, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	res, err := vm.initClient.Exec(timeoutCtx, api.ExecOptions{
		Cmd:       cmd,
		TimeoutMs: int(timeout.Milliseconds()),
	})
	if err != nil {
		return nil, err
	}

	return &api.ExecResult{
		Stderr:   res.Stderr,
		Stdout:   res.Stdout,
		ExitCode: int(res.ExitCode),
	}, nil
}
