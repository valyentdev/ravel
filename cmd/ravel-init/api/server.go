package api

import (
	"os/exec"

	"github.com/valyentdev/ravel/internal/vminit"
	"github.com/valyentdev/ravel/pkg/proto"
	"google.golang.org/grpc"
)

type ExitEvent struct {
	ExitCode int64
}
type Server struct {
	cmd     *exec.Cmd
	updates chan struct{}
	status  *proto.InitStatus
	server  *grpc.Server

	config vminit.Config
}

var _ proto.InitServiceServer = (*Server)(nil)

func New(config vminit.Config, cmd *exec.Cmd) *Server {
	grpcServer := grpc.NewServer()
	server := &Server{
		updates: make(chan struct{}, 1),
		config:  config,
		cmd:     cmd,
		status:  &proto.InitStatus{},
		server:  grpcServer,
	}

	proto.RegisterInitServiceServer(grpcServer, server)

	return server
}
