package manager

import (
	"context"
	"time"

	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/proto"
	"github.com/valyentdev/ravel/pkg/ravelerrors"
	"golang.org/x/exp/slog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (m *Manager) ListMachines(ctx context.Context, namespace string, fleetId string, includeDestroyed bool) ([]core.Machine, error) {
	machines, err := m.clusterState.ListAPIMachines(ctx, namespace, fleetId)
	if err != nil {
		slog.Error("Failed to list machines", "error", err)
		return nil, ravelerrors.NewUnknown("Failed to list machines")
	}

	return machines, nil

}

func (m *Manager) GetMachine(ctx context.Context, namespace string, fleetId string, id string) (core.Machine, error) {
	return m.clusterState.GetAPIMachine(ctx, namespace, fleetId, id)
}

func (m *Manager) DestroyMachine(ctx context.Context, namespace string, fleetId string, id string) error {
	machine, err := m.clusterState.GetMachine(ctx, namespace, fleetId, id, false)
	if err != nil {
		return err
	}

	conn, agentClient, err := m.getAgentClient(machine.Node)
	if err != nil {
		return err
	}

	defer conn.Close()

	_, err = agentClient.DestroyInstance(ctx, &proto.DestroyInstanceRequest{
		Id: machine.InstanceId,
	})
	if err != nil {
		return ravelerrors.NewUnknown("Failed to delete machine")
	}

	return nil
}

func (m *Manager) StartMachine(ctx context.Context, namespace string, fleetId string, id string) error {
	apiMachine, err := m.GetMachine(ctx, namespace, fleetId, id)
	if err != nil {
		slog.Error("Failed to get machine", "error", err)
		return err
	}

	if apiMachine.State == core.MachineStatusDestroyed {
		return ravelerrors.NewFailedPrecondition("machine is destroyed")
	}

	instance, err := m.clusterState.GetCurrentMachineInstance(ctx, id)
	if err != nil {
		return err
	}

	conn, agentClient, err := m.getAgentClient(instance.Instance.Node)
	if err != nil {
		slog.Error("Failed to get agent client", "error", err)
		return err
	}
	defer conn.Close()

	_, err = agentClient.StartInstance(ctx, &proto.StartInstanceRequest{
		Id: instance.Instance.Id,
	})

	if err != nil {
		slog.Error("Failed to start machine", "error", err)
		return ravelerrors.NewUnknown("Failed to start machine")
	}

	return nil
}

func (m *Manager) StopMachine(ctx context.Context, namespace string, fleetId string, id string) error {
	apiMachine, err := m.GetMachine(ctx, namespace, fleetId, id)
	if err != nil {
		return err
	}

	if apiMachine.State == core.MachineStatusDestroyed {
		return ravelerrors.NewFailedPrecondition("machine is destroyed")
	}

	instance, err := m.clusterState.GetCurrentMachineInstance(ctx, id)
	if err != nil {
		return err
	}

	conn, agentClient, err := m.getAgentClient(instance.Instance.Node)
	if err != nil {
		return err
	}

	defer conn.Close()

	_, err = agentClient.StopInstance(ctx, &proto.StopInstanceRequest{
		Id: instance.Instance.Id,
	})
	if err != nil {
		return ravelerrors.NewUnknown("Failed to stop machine")
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

func (m *Manager) WaitMachine(
	ctx context.Context,
	namespace string,
	fleetId string,
	id string,
	state core.MachineStatus,
	opts ...WaitOpt,
) error {
	machine, err := m.GetMachine(ctx, namespace, fleetId, id)
	if err != nil {
		return err
	}

	opt := &waitOpt{
		instanceId: machine.InstanceId,
		timeout:    time.Second * 30,
	}

	timeoutCtx, cancelTimeoutCtx := context.WithTimeout(ctx, opt.timeout)
	defer cancelTimeoutCtx()

	for _, o := range opts {
		o(opt)
	}
	cancel, updates, err := m.clusterState.WatchInstance(ctx, id, opt.instanceId)
	if err != nil {
		return err
	}
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			return ravelerrors.NewDeadlineExceeded("timeout reached while waiting for machine status")
		case update := <-updates:
			if update.Status.Status == state {
				return nil
			}
		}
	}

}

func (m *Manager) getAgentClient(node string) (*grpc.ClientConn, proto.AgentServiceClient, error) {
	member, err := m.clusterState.GetNode(context.Background(), node)
	if err != nil {
		return nil, nil, ravelerrors.NewUnknown("host not found")
	}

	conn, err := grpc.NewClient(member.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, ravelerrors.NewUnavailable("Failed to connect to agent")
	}

	agentClient := proto.NewAgentServiceClient(conn)

	return conn, agentClient, nil
}
