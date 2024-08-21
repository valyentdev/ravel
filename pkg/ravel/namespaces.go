package ravel

import (
	"context"
	"time"

	"github.com/valyentdev/ravel/pkg/core"
)

func (r *Ravel) GetNamespace(ctx context.Context, name string) (*core.Namespace, error) {
	namespace, err := r.db.GetNamespace(ctx, name)
	if err != nil {
		return nil, err
	}

	return namespace, nil
}

func (r *Ravel) CreateNamespace(ctx context.Context, name string) (*core.Namespace, error) {
	err := validateObjectName(name)
	if err != nil {
		return nil, err
	}

	namespace := core.Namespace{
		Name:      name,
		CreatedAt: time.Now(),
	}
	err = r.db.CreateNamespace(ctx, namespace)
	if err != nil {
		return nil, err
	}

	return &namespace, nil
}

func (r *Ravel) ListNamespaces(ctx context.Context) ([]core.Namespace, error) {
	namespaces, err := r.db.ListNamespaces(ctx)
	if err != nil {
		return nil, err
	}

	return namespaces, nil
}

func (r *Ravel) DeleteNamespace(ctx context.Context, name string) error {
	namespace, err := r.GetNamespace(ctx, name)
	if err != nil {
		return err
	}
	err = r.db.DestroyNamespace(ctx, namespace.Name)
	if err != nil {
		return err
	}
	return nil
}
