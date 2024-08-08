package api

import (
	"context"
	"os/exec"

	"github.com/valyentdev/firecracker-go-sdk/vsock"
	"github.com/valyentdev/ravel/pkg/proto"
	"golang.org/x/sys/unix"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) Follow(req *emptypb.Empty, server proto.InitService_FollowServer) error {
	server.Send(s.status)
	for range s.updates {
		if err := server.Send(s.status); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) Signal(ctx context.Context, req *proto.SignalRequest) (*emptypb.Empty, error) {
	signal := unix.Signal(req.Signal)
	if err := s.cmd.Process.Signal(signal); err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *Server) Serve() {
	listener, err := vsock.Listener(context.Background(), nil, 10000)
	if err != nil {
		panic(err)
	}
	if err := s.server.Serve(listener); err != nil {
		panic(err)
	}
}

func (s *Server) UpdateStatus(status *proto.InitStatus) {
	s.status = status
	if len(s.updates) == 0 {
		s.updates <- struct{}{}
	}

}

func (s Server) Exec(ctx context.Context, request *proto.ExecRequest) (*proto.ExecResponse, error) {
	cmdPath, err := exec.LookPath(request.Cmd[0])
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, cmdPath, request.Cmd[1:]...)
	cmd.Env = append(append(s.config.ImageConfig.Env, s.config.ExtraEnv...), request.Env...)
	workingDir := "/"
	if s.config.ImageConfig.WorkingDir != nil {
		workingDir = *s.config.ImageConfig.WorkingDir
	}
	cmd.Dir = workingDir
	if request.WorkingDir != nil {
		cmd.Dir = *request.WorkingDir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return &proto.ExecResponse{
		ExitCode: int32(cmd.ProcessState.ExitCode()),
		Output:   output,
	}, nil
}

func (s *Server) HealthCheck(context.Context, *emptypb.Empty) (*proto.InitStatus, error) {
	return s.status, nil
}
