package state

import (
	"context"
	"log/slog"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/errdefs"
)

func (s *State) CreateNamespace(ctx context.Context, namespace api.Namespace) error {
	return s.db.CreateNamespace(ctx, namespace)
}

func (s *State) ListNamespaces(ctx context.Context) ([]api.Namespace, error) {
	return s.db.ListNamespaces(ctx)
}

func (s *State) GetNamespace(ctx context.Context, name string) (*api.Namespace, error) {
	return s.db.GetNamespace(ctx, name)
}

func (q *State) DestroyNamespace(ctx context.Context, name string) error {
	tx, err := q.db.BeginTx(ctx)
	if err != nil {
		return err
	}

	_, err = q.db.GetNamespaceForUpdate(ctx, name)
	if err != nil {
		return err
	}

	activeFleets, err := tx.CountActiveFleets(ctx, name)
	if err != nil {
		return err
	}

	if activeFleets > 0 {
		return errdefs.NewFailedPrecondition("namespace still has active fleets")
	}

	err = tx.DestroyNamespace(ctx, name)
	if err != nil {
		return err
	}

	err = q.clusterState.DestroyNamespaceData(ctx, name)
	if err != nil {
		slog.Error("failed to destroy namespace data", "namespace", name, "err", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
