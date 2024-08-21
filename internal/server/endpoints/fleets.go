package endpoints

import (
	"context"

	"github.com/valyentdev/ravel/pkg/ravel"
)

type CreateFleetBody struct {
	Name string `json:"name"`
}

type CreateFleetRequest struct {
	NSResolver
	Body *CreateFleetBody
}

type CreateFleetResponse struct {
	Body *ravel.Fleet `json:"fleet"`
}

func (e *Endpoints) createFleet(ctx context.Context, req *CreateFleetRequest) (*CreateFleetResponse, error) {
	fleet, err := e.ravel.CreateFleet(ctx, req.Namespace, req.Body.Name)
	if err != nil {
		e.log("Failed to create fleet", err)
		return nil, err
	}

	return &CreateFleetResponse{
		Body: fleet,
	}, nil
}

type ListFleetsRequest struct {
	NSResolver
}

type ListFleetsResponse struct {
	Body []ravel.Fleet `json:"fleets"`
}

func (e *Endpoints) listFleets(ctx context.Context, req *ListFleetsRequest) (*ListFleetsResponse, error) {
	fleets, err := e.ravel.ListFleets(ctx, req.Namespace)
	if err != nil {
		e.log("Failed to list fleets", err)
		return nil, err
	}
	return &ListFleetsResponse{Body: fleets}, nil
}

type GetFleetRequest struct {
	FleetResolver
}

type GetFleetResponse struct {
	Body *ravel.Fleet `json:"fleet"`
}

func (e *Endpoints) getFleet(ctx context.Context, req *GetFleetRequest) (*GetFleetResponse, error) {
	fleet, err := e.ravel.GetFleet(ctx, req.Namespace, req.Fleet)
	if err != nil {
		e.log("Failed to get fleet", err)
		return nil, err
	}

	return &GetFleetResponse{Body: fleet}, nil
}

type DestroyFleetRequest struct {
	FleetResolver
}

type DestroyFleetResponse struct {
}

func (e *Endpoints) destroyFleet(ctx context.Context, req *DestroyFleetRequest) (*DestroyFleetResponse, error) {
	err := e.ravel.DestroyFleet(ctx, req.Namespace, req.Fleet)
	if err != nil {
		e.log("Failed to destroy fleet", err)
		return nil, err
	}

	return nil, nil
}
