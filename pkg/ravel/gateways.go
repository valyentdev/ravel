package ravel

import (
	"context"
	"strings"

	"github.com/valyentdev/ravel/internal/id"
	"github.com/valyentdev/ravel/pkg/core"
)

type Gateway = core.Gateway

func (r *Ravel) GetGateway(ctx context.Context, namespace string, gateway string) (Gateway, error) {
	var g Gateway
	var err error

	if strings.HasPrefix(gateway, "gw_") {
		g, err = r.db.GetGatewayById(ctx, namespace, gateway)
		if err != nil {
			return Gateway{}, err
		}
	} else {
		g, err = r.db.GetGatewayByName(ctx, namespace, gateway)
		if err != nil {
			return Gateway{}, err
		}
	}

	return g, nil
}

func (r *Ravel) ListGateways(ctx context.Context, namespace string) ([]Gateway, error) {
	_, err := r.GetNamespace(ctx, namespace)
	if err != nil {
		return nil, err
	}

	return r.db.ListGateways(ctx, namespace)
}

type CreateGatewayOptions struct {
	Name       string `json:"name"`
	Fleet      string `json:"fleet"`
	TargetPort int    `json:"target_port"`
}

func (r *Ravel) CreateGateway(ctx context.Context, namespace string, options CreateGatewayOptions) (Gateway, error) {
	err := validateObjectName(options.Name)
	if err != nil {
		return Gateway{}, err
	}

	f, err := r.GetFleet(ctx, namespace, options.Fleet)
	if err != nil {
		return Gateway{}, err
	}

	gateway := Gateway{
		Id:         id.GeneratePrefixed("gw"),
		Name:       options.Name,
		Namespace:  namespace,
		FleetId:    f.Id,
		TargetPort: options.TargetPort,
	}

	err = r.state.CreateGateway(gateway)
	if err != nil {
		return Gateway{}, err
	}

	return gateway, nil
}

func (r *Ravel) DeleteGateway(ctx context.Context, namespace string, gateway string) error {
	g, err := r.GetGateway(ctx, namespace, gateway)
	if err != nil {
		return err
	}

	return r.state.DeleteGateway(ctx, g.Id)
}
