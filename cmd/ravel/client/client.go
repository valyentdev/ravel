package raveldclient

import (
	"github.com/valyentdev/ravel/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var DaemonClient proto.AgentServiceClient

func init() {
	conn, err := grpc.NewClient("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	DaemonClient = proto.NewAgentServiceClient(conn)

}
