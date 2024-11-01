package endpoints

import (
	"context"

	"github.com/valyentdev/ravel/pkg/ravel"
)

type CreateGatewayRequest struct {
	NSResolver
	Body ravel.CreateGatewayOptions
}

type CreateGatewayResponse struct {
	Body ravel.Gateway `json:"gateway"`
}

func (e *Endpoints) createGateway(ctx context.Context, req *CreateGatewayRequest) (*CreateGatewayResponse, error) {
	gateway, err := e.ravel.CreateGateway(ctx, req.Namespace, req.Body)
	if err != nil {
		e.log("Failed to create gateway", err)
		return nil, err
	}

	return &CreateGatewayResponse{
		Body: gateway,
	}, nil
}

type ListGatewaysRequest struct {
	NSResolver
}

type ListGatewaysResponse struct {
	Body []ravel.Gateway `json:"gateways"`
}

func (e *Endpoints) listGateways(ctx context.Context, req *ListGatewaysRequest) (*ListGatewaysResponse, error) {
	gateways, err := e.ravel.ListGateways(ctx, req.Namespace)
	if err != nil {
		e.log("Failed to list gateways", err)
		return nil, err
	}
	return &ListGatewaysResponse{Body: gateways}, nil
}

type GetGatewayRequest struct {
	NSResolver
	Gateway string `path:"gateway"`
}

type GetGatewayResponse struct {
	Body ravel.Gateway `json:"gateway"`
}

func (e *Endpoints) getGateway(ctx context.Context, req *GetGatewayRequest) (*GetGatewayResponse, error) {
	gateway, err := e.ravel.GetGateway(ctx, req.Namespace, req.Gateway)
	if err != nil {
		e.log("Failed to get gateway", err)
		return nil, err
	}

	return &GetGatewayResponse{Body: gateway}, nil
}

type DestroyGatewayRequest struct {
	NSResolver
	Gateway string `path:"gateway"`
}

type DestroyGatewayResponse struct {
}

func (e *Endpoints) destroyGateway(ctx context.Context, req *DestroyGatewayRequest) (*DestroyGatewayResponse, error) {
	err := e.ravel.DeleteGateway(ctx, req.Namespace, req.Gateway)
	if err != nil {
		e.log("Failed to destroy gateway", err)
		return nil, err
	}

	return nil, nil
}
