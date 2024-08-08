package agent

import (
	"context"
	"errors"
	"time"

	"github.com/valyentdev/ravel/internal/agent/instance"
	"github.com/valyentdev/ravel/pkg/proto"
	"github.com/valyentdev/ravel/pkg/ravelerrors"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Agent) StartInstance(ctx context.Context, request *proto.StartInstanceRequest) (*proto.StartInstanceResponse, error) {
	instance, err := s.getInstance(request.GetId())
	if err != nil {
		return nil, err
	}

	err = instance.Start(context.Background())
	if err != nil {
		return nil, err
	}

	return &proto.StartInstanceResponse{
		Id: request.GetId(),
	}, nil
}

func (s *Agent) DestroyInstance(ctx context.Context, request *proto.DestroyInstanceRequest) (*emptypb.Empty, error) {
	id := request.GetId()
	if id == "" {
		return nil, errors.New("id is required")
	}

	instance, err := s.getInstance(id)
	if err != nil {
		return nil, err
	}

	err = instance.Destroy(context.Background(), false)
	if err != nil {
		return nil, err
	}

	s.lock.Lock()
	delete(s.instances, id)
	s.lock.Unlock()

	return &emptypb.Empty{}, nil
}

func (s *Agent) StopInstance(ctx context.Context, request *proto.StopInstanceRequest) (*proto.StopInstanceResponse, error) {
	instance, err := s.getInstance(request.GetId())
	if err != nil {
		return nil, err
	}

	err = instance.Stop(context.Background(), "", 5*time.Second)
	if err != nil {
		return nil, err
	}

	return &proto.StopInstanceResponse{
		Id: request.GetId(),
	}, nil
}

func (s *Agent) InstanceExec(ctx context.Context, request *proto.InstanceExecRequest) (*proto.InstanceExecResponse, error) {
	instance, err := s.getInstance(request.GetInstanceId())
	if err != nil {
		return nil, err
	}

	res, err := instance.Exec(ctx, request.ExecRequest.Cmd, time.Duration(request.Timeout)*time.Millisecond)
	if err != nil {
		return nil, err
	}

	return &proto.InstanceExecResponse{
		Output: res.Stdout,
	}, nil
}

func (d *Agent) ListInstances(ctx context.Context, req *proto.ListInstancesRequest) (*proto.ListInstancesResponse, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	instances := []*proto.Instance{}
	for _, m := range d.instances {
		i := m.Instance()
		status := m.Status()

		p := InstanceToProto(&i)

		p.State = string(status)

		instances = append(instances, p)

	}

	return &proto.ListInstancesResponse{
		Instances: instances,
	}, nil
}

func (a *Agent) GetInstance(ctx context.Context, req *proto.GetInstanceRequest) (*proto.GetInstanceResponse, error) {
	m, err := a.getInstance(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	i := m.Instance()

	return &proto.GetInstanceResponse{
		Instance: InstanceToProto(&i),
	}, nil
}

func (a *Agent) getInstance(id string) (*instance.Manager, error) {
	a.lock.RLock()
	m, ok := a.instances[id]
	a.lock.RUnlock()
	if !ok {
		return nil, ravelerrors.ErrInstanceNotFound
	}

	return m, nil
}
