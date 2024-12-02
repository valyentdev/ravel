package ravel

import (
	"context"
	"time"

	"github.com/valyentdev/ravel/api"
)

func (r *Ravel) GetNamespace(ctx context.Context, name string) (*api.Namespace, error) {
	namespace, err := r.state.GetNamespace(ctx, name)
	if err != nil {
		return nil, err
	}

	return namespace, nil
}

func (r *Ravel) CreateNamespace(ctx context.Context, name string) (*api.Namespace, error) {
	err := validateObjectName(name)
	if err != nil {
		return nil, err
	}

	namespace := api.Namespace{
		Name:      name,
		CreatedAt: time.Now(),
	}
	err = r.state.CreateNamespace(ctx, namespace)
	if err != nil {
		return nil, err
	}

	return &namespace, nil
}

func (r *Ravel) ListNamespaces(ctx context.Context) ([]api.Namespace, error) {
	namespaces, err := r.state.ListNamespaces(ctx)
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
	err = r.state.DestroyNamespace(ctx, namespace.Name)
	if err != nil {
		return err
	}
	return nil
}
