package orchestrator

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/core/api"
)

func (m *Orchestrator) DestroyMachine(ctx context.Context, machine core.Machine, force bool) error {
	agentClient, err := m.getAgentClient(machine.Node)
	if err != nil {
		return err
	}

	err = agentClient.DestroyInstance(ctx, machine.InstanceId, force)
	if err != nil {
		return err
	}

	return nil
}

func (m *Orchestrator) StartMachineInstance(ctx context.Context, machine core.Machine) error {
	agentClient, err := m.getAgentClient(machine.Node)
	if err != nil {
		return err
	}

	err = agentClient.StartInstance(ctx, machine.InstanceId)
	if err != nil {
		return core.NewUnknown("Failed to start machine")
	}

	return nil
}

func (m *Orchestrator) StopMachineInstance(ctx context.Context, machine core.Machine, stopConfig *core.StopConfig) error {
	agentClient, err := m.getAgentClient(machine.Node)
	if err != nil {
		return err
	}

	err = agentClient.StopInstance(ctx, machine.InstanceId, stopConfig)
	if err != nil {
		return core.NewUnknown("Failed to stop machine")
	}

	return nil
}

type waitOpt struct {
	instanceId string
	timeout    time.Duration
}

type WaitOpt func(*waitOpt)

func WithInstanceId(instanceId string) WaitOpt {
	return func(o *waitOpt) {
		o.instanceId = instanceId
	}
}

func WithTimeout(timeout time.Duration) WaitOpt {
	return func(o *waitOpt) {
		o.timeout = timeout
	}
}

func (m *Orchestrator) WaitMachine(
	ctx context.Context,
	machine core.Machine,
	state core.MachineStatus,
	opts ...WaitOpt,
) error {
	opt := &waitOpt{
		instanceId: machine.InstanceId,
		timeout:    time.Second * 30,
	}

	timeoutCtx, cancelTimeoutCtx := context.WithTimeout(ctx, opt.timeout)
	defer cancelTimeoutCtx()

	for _, o := range opts {
		o(opt)
	}
	cancel, updates, err := m.clusterState.WatchInstance(ctx, machine.Id, opt.instanceId)
	if err != nil {
		return err
	}
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			return core.NewDeadlineExceeded("timeout reached while waiting for machine status")
		case update := <-updates:
			if update.Status == state {
				return nil
			}
		}
	}

}

func (m *Orchestrator) GetMachineLogsRaw(ctx context.Context, machine core.Machine, follow bool) (io.ReadCloser, error) {
	agentClient, err := m.getAgentClient(machine.Node)
	if err != nil {
		return nil, err
	}

	return agentClient.GetInstanceLogsRaw(ctx, machine.InstanceId, follow)
}

func (m *Orchestrator) getAgentClient(node string) (*api.AgentClient, error) {
	member, err := m.clusterState.GetNode(context.Background(), node)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	client := api.NewAgentClient(m.httpClient, fmt.Sprintf("http://%s", member.Address))

	return client, nil
}
