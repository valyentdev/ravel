package agent

import (
	"context"
	"time"

	"github.com/valyentdev/ravel/internal/agent/instance"
	"github.com/valyentdev/ravel/pkg/core"
)

func (s *Agent) StartInstance(ctx context.Context, id string) error {
	instance, err := s.getInstance(id)
	if err != nil {
		return err
	}

	err = instance.Start(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (s *Agent) DestroyInstance(ctx context.Context, id string, force bool) error {
	instance, err := s.getInstance(id)
	if err != nil {
		return err
	}

	err = instance.Destroy(context.Background(), force)
	if err != nil {
		return err
	}

	err = s.reservations.DeleteReservation(instance.Instance().MachineId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Agent) StopInstance(ctx context.Context, id string, opt *core.StopConfig) error {
	instance, err := s.getInstance(id)
	if err != nil {
		return nil
	}

	instanceStopConfig := instance.Instance().Config.StopConfig

	var timeout time.Duration

	if opt != nil && opt.Timeout != nil {
		timeout = opt.GetTimeout()
	} else {
		timeout = instanceStopConfig.GetTimeout()
	}

	var signal string
	if opt != nil && opt.Signal != nil {
		signal = opt.GetSignal()
	} else {
		signal = instanceStopConfig.GetSignal()
	}

	err = instance.Stop(context.Background(), signal, timeout)
	if err != nil {
		return nil
	}

	return nil
}

func (s *Agent) InstanceExec(ctx context.Context, id string, opt core.InstanceExecOptions) (*core.ExecResult, error) {
	instance, err := s.getInstance(id)
	if err != nil {
		return nil, err
	}

	if len(opt.Cmd) == 0 {
		return nil, core.NewInvalidArgument("cmd is required")
	}

	res, err := instance.Exec(ctx, opt.Cmd, opt.GetTimeout())
	if err != nil {
		return nil, err
	}

	return &core.ExecResult{
		Stdout:   string(res.Stdout),
		ExitCode: res.ExitCode,
	}, nil
}

func (d *Agent) ListInstances(ctx context.Context) ([]core.Instance, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	instances := []core.Instance{}
	for _, m := range d.instances {
		i := m.Instance()
		instances = append(instances, i)
	}

	return instances, nil
}

func (a *Agent) GetInstance(ctx context.Context, id string) (*core.Instance, error) {
	m, err := a.getInstance(id)
	if err != nil {
		return nil, err
	}
	i := m.Instance()

	return &i, nil
}

func (a *Agent) getInstance(id string) (*instance.Manager, error) {
	a.lock.RLock()
	m, ok := a.instances[id]
	a.lock.RUnlock()
	if !ok {
		return nil, core.NewNotFound("instance not found")
	}

	return m, nil
}

func (a *Agent) SubscribeToInstanceLogs(ctx context.Context, id string) (<-chan []*core.LogEntry, error) {
	m, err := a.getInstance(id)
	if err != nil {
		return nil, err
	}

	sub := m.SubscribeToLogs()

	ch := sub.Ch()

	go func() {
		<-ctx.Done()
		sub.Unsubscribe()
	}()

	return ch, nil
}

func (a *Agent) GetInstanceLogs(ctx context.Context, id string) ([]*core.LogEntry, error) {
	m, err := a.getInstance(id)
	if err != nil {
		return nil, err
	}

	return m.GetLog(), nil
}
