package ravel

import (
	"context"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/internal/id"
)

type Gateway = api.Gateway

func (r *Ravel) GetGateway(ctx context.Context, namespace string, gateway string) (Gateway, error) {
	ns, err := r.GetNamespace(ctx, namespace)
	if err != nil {
		return Gateway{}, err
	}

	return r.state.GetGateway(ctx, ns.Name, gateway)
}

type listGatewaysOptions struct {
	fleet string
}

type ListGatewaysOpt func(*listGatewaysOptions)

func WithFleet(fleet string) ListGatewaysOpt {
	return func(o *listGatewaysOptions) {
		o.fleet = fleet
	}
}

func (r *Ravel) ListGateways(ctx context.Context, namespace string, opts ...ListGatewaysOpt) ([]Gateway, error) {
	_, err := r.GetNamespace(ctx, namespace)
	if err != nil {
		return nil, err
	}

	var options listGatewaysOptions
	for _, opt := range opts {
		opt(&options)
	}

	if options.fleet != "" {
		f, err := r.GetFleet(ctx, namespace, options.fleet)
		if err != nil {
			return nil, err
		}

		return r.state.ListGatewaysOnFleet(ctx, namespace, f.Id)
	}

	return r.state.ListGateways(ctx, namespace)
}

type CreateGatewayOptions = api.CreateGatewayPayload

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
