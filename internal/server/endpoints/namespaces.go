package endpoints

import (
	"context"

	"github.com/valyentdev/ravel/pkg/ravel"
)

type CreateNamespaceBody struct {
	Name string `json:"name"`
}

type CreateNamespaceRequest struct {
	Body *CreateNamespaceBody
}

type CreateNamespaceResponse struct {
	Body *ravel.Namespace `json:"namespace"`
}

func (e *Endpoints) createNamespace(ctx context.Context, req *CreateNamespaceRequest) (*CreateNamespaceResponse, error) {
	ns, err := e.ravel.CreateNamespace(ctx, req.Body.Name)
	if err != nil {
		e.log("Failed to create namespace", err)
		return nil, err
	}

	return &CreateNamespaceResponse{
		Body: ns,
	}, nil
}

type ListNamespacesRequest struct {
}

type ListNamespacesResponse struct {
	Body []ravel.Namespace `json:"namespaces"`
}

func (e *Endpoints) listNamespaces(ctx context.Context, _ *ListNamespacesRequest) (*ListNamespacesResponse, error) {
	namespaces, err := e.ravel.ListNamespaces(ctx)
	if err != nil {
		e.log("Failed to list namespaces", err)
		return nil, err
	}
	return &ListNamespacesResponse{Body: namespaces}, nil
}

type GetNamespaceRequest struct {
	Namespace string `path:"namespace"`
}

type GetNamespaceResponse struct {
	Body *ravel.Namespace `json:"namespace"`
}

func (e *Endpoints) getNamespace(ctx context.Context, req *GetNamespaceRequest) (*GetNamespaceResponse, error) {
	ns, err := e.ravel.GetNamespace(ctx, req.Namespace)
	if err != nil {
		e.log("Failed to get namespace", err)
		return nil, err
	}
	return &GetNamespaceResponse{Body: ns}, nil
}

type DestroyNamespaceRequest struct {
	Namespace string `path:"namespace"`
}

type DestroyNamespaceResponse struct {
}

func (e *Endpoints) destroyNamespace(ctx context.Context, req *GetNamespaceRequest) (*DestroyNamespaceResponse, error) {
	err := e.ravel.DeleteNamespace(ctx, req.Namespace)
	if err != nil {
		e.log("Failed to destroy namespace", err)
		return nil, err
	}
	return &DestroyNamespaceResponse{}, nil
}
