package ravel

import (
	"context"

	"github.com/valyentdev/ravel/api"
)

func (r *Ravel) ListNodes(ctx context.Context) ([]api.Node, error) {
	nodes, err := r.s.ListNodes(ctx)
	if err != nil {
		return nil, err
	}

	return nodes, nil
}
