package vminit

import (
	"os/exec"

	"github.com/valyentdev/ravel/pkg/runtimes/container/vminit/proto"
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

	config Config
}

var _ proto.InitServiceServer = (*Server)(nil)

func newInitAPI(config Config, cmd *exec.Cmd) *Server {
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
