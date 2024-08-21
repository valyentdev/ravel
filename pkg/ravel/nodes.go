package ravel

import (
	"context"

	"github.com/valyentdev/ravel/pkg/core"
)

func (r *Ravel) ListNodes(ctx context.Context) ([]core.Node, error) {
	nodes, err := r.o.ListNodes(ctx)
	if err != nil {
		return nil, err
	}

	return nodes, nil
}
