package instancerunner

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/api/errdefs"
	"github.com/valyentdev/ravel/core/instance"
	"github.com/valyentdev/ravel/pkg/pubsub"
	"github.com/valyentdev/ravel/runtime/logging"
)

type InstanceRunner struct {
	logger *logging.InstanceLogger

	store instance.InstanceStore
	mutex sync.Mutex

	instance      instance.Instance
	stateObserver *pubsub.Observable[instance.State]
	instanceLock  sync.RWMutex

	networking  instance.NetworkingService
	vmBuilder   instance.Builder
	runnerMutex sync.Mutex
	runner      *vmRunner
	waitForExit []chan instance.ExitResult
}

var errNotRunning = errdefs.NewFailedPrecondition("instance is not running")

func (ir *InstanceRunner) newVMRunner() *vmRunner {
	return newVMRunner(ir.Instance(), ir.logger, ir.vmBuilder, ir.networking)
}

func (ir *InstanceRunner) getVMRunner() *vmRunner {
	ir.runnerMutex.Lock()
	defer ir.runnerMutex.Unlock()
	return ir.runner
}

func (ir *InstanceRunner) setVMRunner(runner *vmRunner) {
	ir.runnerMutex.Lock()
	defer ir.runnerMutex.Unlock()
	ir.runner = runner
}

func New(
	store instance.InstanceStore,
	instance instance.Instance,
	ns instance.NetworkingService,
	vmBuilder instance.Builder,
) *InstanceRunner {
	return &InstanceRunner{
		logger:        logging.NewInstanceLogger(instance.Id),
		store:         store,
		instance:      instance,
		networking:    ns,
		stateObserver: pubsub.NewObservable(instance.State),
		vmBuilder:     vmBuilder,
	}
}

const defaultExecTimeout = time.Duration(5) * time.Second

func (ir *InstanceRunner) Exec(ctx context.Context, cmd []string, timeout time.Duration) (*api.ExecResult, error) {
	if len(cmd) == 0 {
		return nil, errdefs.NewInvalidArgument("cmd is required")
	}

	if timeout == 0 {
		timeout = defaultExecTimeout
	}

	if timeout > 30*time.Second {
		return nil, errdefs.NewInvalidArgument("timeout must be less than 30 seconds")
	}

	runner := ir.getVMRunner()
	if runner == nil {
		return nil, errNotRunning
	}

	return runner.Exec(ctx, cmd, timeout)
}

func (ir *InstanceRunner) GetLog() []*api.LogEntry {
	return ir.logger.GetLog()
}

func (ir *InstanceRunner) SubscribeToLogs() ([]*api.LogEntry, *logging.LogSubscriber) {
	return ir.logger.Subscribe()
}

func (ir *InstanceRunner) Signal(ctx context.Context, signal string) error {
	runner := ir.getVMRunner()
	if runner == nil {
		return errNotRunning
	}

	return runner.Signal(ctx, signal)
}

func (s *InstanceRunner) Instance() instance.Instance {
	s.instanceLock.RLock()
	defer s.instanceLock.RUnlock()
	return s.instance
}

func (s *InstanceRunner) Status() instance.InstanceStatus {
	return s.Instance().State.Status
}

func (s *InstanceRunner) lock() {
	s.mutex.Lock()
}

func (s *InstanceRunner) unlock() {
	s.mutex.Unlock()
}

func (s *InstanceRunner) updateInstanceStateFunc(f func(state *instance.State)) error {
	s.instanceLock.Lock()
	defer s.instanceLock.Unlock()

	f(&s.instance.State)
	err := s.store.PutInstance(s.instance)
	if err != nil {
		slog.Error("failed to update instance state", "error", err)
		return err
	}

	s.stateObserver.Set(s.instance.State)
	return nil
}

func (s *InstanceRunner) updateInstanceState(state instance.State) error {
	s.instanceLock.Lock()
	defer s.instanceLock.Unlock()
	s.instance.State = state
	err := s.store.PutInstance(s.instance)
	if err != nil {
		slog.Error("failed to update instance state", "error", err)
		return err
	}

	s.stateObserver.Set(state)
	return nil
}
