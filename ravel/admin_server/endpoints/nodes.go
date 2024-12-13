package endpoints

import (
	"context"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/valyentdev/ravel/api"
)

type ListNodesRequest struct {
}

type ListNodesResponse struct {
	Body []api.Node `json:"nodes"`
}

func (e *Endpoints) listNodes(ctx context.Context, _ *ListNodesRequest) (*ListNodesResponse, error) {
	nodes, err := e.ravel.ListNodes(ctx)
	if err != nil {
		slog.Error("Failed to list nodes", "error", err)
		return nil, huma.NewError(500, "Failed to list nodes")
	}
	return &ListNodesResponse{Body: nodes}, nil

}
