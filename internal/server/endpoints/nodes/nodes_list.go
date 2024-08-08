package nodes

import (
	"context"

	"github.com/valyentdev/ravel/pkg/core"
)

type ListNodesRequest struct {
}

type ListNodesResponse struct {
	Body []core.Node `json:"nodes"`
}

func (e *Endpoint) listNodes(ctx context.Context, _ *ListNodesRequest) (*ListNodesResponse, error) {
	nodes, err := e.m.ListNodes(ctx)
	if err != nil {
		return nil, err
	}

	return &ListNodesResponse{Body: nodes}, nil

}
