package api

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/valyentdev/ravel/initd"
	"github.com/valyentdev/ravel/initd/environment"
)

type InternalEndpoint struct {
	env *environment.Env
}

func (e *InternalEndpoint) registerRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		Method:      "GET",
		Path:        "/wait",
		OperationID: "waitExit",
		Description: "Wait for the container main process to exit",
	}, e.wait)

	huma.Register(api, huma.Operation{
		Path:        "/signal",
		Method:      "POST",
		OperationID: "signal",
		Description: "Send a signal to the container main process",
	}, e.signal)
}

type WaitRequest struct{}

type WaitResponse struct {
	Body initd.WaitResult
}

func (e *InternalEndpoint) wait(ctx context.Context, req *WaitRequest) (*WaitResponse, error) {
	return &WaitResponse{Body: e.env.Wait()}, nil
}

type SignalRequest struct {
	Body initd.SignalOptions
}

type SignalResponse struct{}

func (e *InternalEndpoint) signal(ctx context.Context, req *SignalRequest) (*SignalResponse, error) {
	err := e.env.Signal(req.Body.Signal)
	if err != nil {
		return nil, err
	}
	return &SignalResponse{}, nil
}
