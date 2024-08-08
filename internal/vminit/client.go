package vminit

import (
	"context"
	"net"

	"github.com/valyentdev/firecracker-go-sdk/vsock"
	"github.com/valyentdev/ravel/pkg/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewClient(path string) (*grpc.ClientConn, proto.InitServiceClient, error) {
	conn, err := grpc.NewClient("localhost", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
		return vsock.Dial(path, 10000)
	}))
	if err != nil {
		return nil, nil, err
	}
	return conn, proto.NewInitServiceClient(conn), nil
}
