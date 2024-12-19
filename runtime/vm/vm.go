package vm

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"sync/atomic"
	"time"

	vminit "github.com/valyentdev/ravel-init/client"
	"github.com/valyentdev/ravel-init/proto"
	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/instance"
	"github.com/valyentdev/ravel/pkg/cloudhypervisor"

	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	unknownExitCode = -1
)

type vm struct {
	cmd                    *exec.Cmd
	successFullyShutdowned atomic.Bool
	id                     string
	runResult              *RunResult
	vmConfig               cloudhypervisor.VmConfig
	vmm                    *cloudhypervisor.VMM
	vsock                  string
	stopRequested          bool
	waitChan               chan struct{}
}

var _ instance.VM = (*vm)(nil)

func (vm *vm) Id() string {
	return vm.id
}

func newVM(id string, cmd *exec.Cmd, vmConfig cloudhypervisor.VmConfig) (*vm, error) {
	slog.Debug("creating new VM", "id", id, "socket", getAPISocketPath(id))
	vmm, err := cloudhypervisor.NewVMMClient(getAPISocketPath(id))
	if err != nil {
		return nil, err
	}

	return &vm{
		id:       id,
		cmd:      cmd,
		vmConfig: vmConfig,
		vmm:      vmm,
		waitChan: make(chan struct{}),
		vsock:    getVsockPath(id),
	}, nil
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
	conn, client, err := vminit.NewClient(vm.vsock)
	if err != nil {
		return fmt.Errorf("failed to create init client: %w", err)
	}
	defer conn.Close()

	_, err = client.Signal(ctx, &proto.SignalRequest{
		Signal: int32(sig),
	})
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
	ExitCode       *int64
	ExitedAt       time.Time
}

func (vm *vm) run() {
	defer close(vm.waitChan)
	defer vm.Shutdown(context.Background())

	result := &RunResult{}

	defer func() {
		result.ExitedAt = time.Now()
		vm.runResult = result
	}()

	conn, initClient, err := vminit.NewClient(vm.vsock)
	if err != nil {
		slog.Error("failed to create init client", "err", err)
		return
	}
	defer conn.Close()

	initStarted := atomic.Bool{}
	processStarted := false

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		time.Sleep(3 * time.Second)
		if !initStarted.Load() {
			cancel()
		}
	}()

	updates, err := initClient.Follow(ctx, &emptypb.Empty{})
	if err != nil {
		slog.Error("failed to follow init", "err", err)
		return
	}

	initStarted.Store(true)

	startedAt := time.Now()

	for {
		update, err := updates.Recv()
		if err != nil {
			result.VMExited = true
			break
		}

		if update.InitFailed {
			result.InitFailed = true
			break
		}

		if update.ProcessExited {
			result.ProcessExited = true
			result.ExitCode = update.ExitCode
			break
		}

		if !processStarted {
			if update.ProcessStarted {
				processStarted = true
			} else if time.Since(startedAt) > 10*time.Second {
				result.InitFailed = true
				break
			}
		}

	}

	if vm.stopRequested {
		result.HasBeenStopped = true
	}

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
		Requested: vm.runResult.HasBeenStopped,
		ExitedAt:  vm.runResult.ExitedAt,
	}

	if vm.runResult.ExitCode != nil {
		er.ExitCode = int(*vm.runResult.ExitCode)
		er.Success = er.ExitCode == 0
	} else {
		er.ExitCode = unknownExitCode
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

	conn, client, err := vminit.NewClient(vm.vsock)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	res, err := client.Exec(timeoutCtx, &proto.ExecRequest{
		Cmd: cmd,
	})
	if err != nil {
		return nil, err
	}

	return &api.ExecResult{
		Stdout:   string(res.Output),
		ExitCode: int(res.ExitCode),
	}, nil
}
