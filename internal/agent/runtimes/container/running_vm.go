package container

import (
	"context"
	"fmt"
	"log/slog"
	"syscall"
	"time"

	"github.com/valyentdev/ravel/internal/agent/logging"
	"github.com/valyentdev/ravel/internal/vminit"
	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/helper/cloudhypervisor"
	"github.com/valyentdev/ravel/pkg/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type runningVM struct {
	runResult     *RunResult
	instance      core.Instance
	vmConfig      cloudhypervisor.VmConfig
	vmm           *cloudhypervisor.VMM
	serial        string
	stopRequested bool
	waitChan      chan struct{}
	logger        *logging.InstanceLogger
}

func (i *runningVM) Id() string {
	return i.instance.Id
}

func NewRunningVM(instance core.Instance, vmConfig cloudhypervisor.VmConfig) (*runningVM, error) {
	vmm, err := cloudhypervisor.NewVMM(
		socketPath(instance.Id),
		cloudhypervisor.WithSysProcAttr(&syscall.SysProcAttr{
			Setsid: true,
		}),
	)
	if err != nil {
		return nil, err
	}

	return &runningVM{
		instance: instance,
		vmConfig: vmConfig,
		vmm:      vmm,
		waitChan: make(chan struct{}),
	}, nil
}

func (i *runningVM) Start() error {
	err := i.vmm.StartVMM(context.Background())
	if err != nil {
		return fmt.Errorf("failed to start vmm for machine %q: %w", i.Id(), err)
	}
	defer func() {
		if err != nil {
			i.vmm.ShutdownVMM(context.Background())
		}
	}()

	err = i.vmm.WaitReady(context.Background())
	if err != nil {
		return fmt.Errorf("failed to wait for vmm to be ready for machine %q: %w", i.Id(), err)
	}

	err = i.vmm.CreateVM(context.Background(), i.vmConfig)
	if err != nil {
		return fmt.Errorf("failed to create vm for machine %q: %w", i.Id(), err)
	}

	err = i.vmm.BootVM(context.Background())
	if err != nil {
		return fmt.Errorf("failed to boot vm for machine %q: %w", i.Id(), err)
	}

	vminfo, err := i.vmm.VMInfo(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get vm info for machine %q: %w", i.Id(), err)
	}

	i.serial = *vminfo.Config.Console.File

	i.logger, err = logging.NewMachineLogger(i.serial)
	if err != nil {
		slog.Error("failed to create logger", "err", err)
	}

	go i.logger.Start(context.Background())

	return nil

}

func (i *runningVM) GetLog() []byte {
	return i.logger.GetLog()
}

func (i *runningVM) SubscribeToLogs(ctx context.Context, ch chan []byte) {
	i.logger.Subscribe(ctx, ch)
}

func (i *runningVM) Signal(ctx context.Context, signal string) error {
	sig := syscallSignal(signal)
	conn, client, err := vminit.NewClient(vsockPath(i.Id()))
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
}

func (i *runningVM) Run() {
	defer close(i.waitChan)
	defer i.vmm.ShutdownVMM(context.Background())

	slog.Info("instance run started", "instance", i.Id())
	conn, initClient, err := vminit.NewClient(vsockPath(i.Id()))
	if err != nil {
		slog.Error("failed to create init client", "err", err)
		return
	}
	defer conn.Close()

	initStarted := false
	processStarted := false
	result := &RunResult{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		time.Sleep(10 * time.Second)
		if !initStarted {
			cancel()
		}
	}()

	updates, err := initClient.Follow(ctx, &emptypb.Empty{})
	if err != nil {
		slog.Error("failed to follow init", "err", err)
		return
	}

	initStarted = true

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

	if i.stopRequested {
		result.HasBeenStopped = true
	}

	i.runResult = result

}

func (i *runningVM) Shutdown(ctx context.Context) error {
	return i.vmm.ShutdownVMM(ctx)
}

func (i *runningVM) Stop(ctx context.Context, signal string) error {
	i.stopRequested = true
	return i.Signal(ctx, signal)
}

func (i *runningVM) WaitExit(ctx context.Context) (exited bool) {
	for {
		select {
		case <-ctx.Done():
			return false
		case <-i.waitChan:
			return true
		}
	}
}

func recoverRunningVM(instance core.Instance) (*runningVM, error) {
	vmm, err := cloudhypervisor.NewVMM(
		socketPath(instance.Id),
		cloudhypervisor.WithSysProcAttr(&syscall.SysProcAttr{
			Setsid: true,
		}),
	)
	if err != nil {
		return nil, err
	}

	i := &runningVM{
		instance: instance,
		vmm:      vmm,
		waitChan: make(chan struct{}),
	}

	err = i.recover()
	if err != nil {
		return nil, err
	}

	return i, nil
}

func (i *runningVM) recover() (err error) {
	state := i.determinateState()
	if !state.isVMMRunning {
		return fmt.Errorf("VMM is not running")
	}
	defer func() {
		if err != nil {
			err = i.vmm.ShutdownVMM(context.Background())
			if err != nil {
				slog.Error("failed to shutdown VMM", "err", err)
			}
		}
	}()

	if !state.isVMRunning {
		return fmt.Errorf("VM is not running")
	}

	serial := state.vminfo.Config.Console.File

	i.serial = *serial

	i.logger, err = logging.NewMachineLogger(i.serial)

	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	return nil
}

type internalState struct {
	isVMMRunning bool
	isVMRunning  bool
	vminfo       *cloudhypervisor.VmInfo
}

func (i *runningVM) determinateState() internalState {
	s := internalState{
		isVMMRunning: false,
		isVMRunning:  false,
	}

	if _, err := i.vmm.PingVMM(context.Background()); err != nil {
		slog.Debug("failed to ping VMM", "err", err)
		return s
	}
	s.isVMMRunning = true

	vminfo, err := i.vmm.VMInfo(context.Background())
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
