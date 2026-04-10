package endpoints

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

type HealthLiveResponse struct {
	Body struct {
		Status string `json:"status"`
	}
}

type HealthReadyResponse struct {
	Body struct {
		Status  string            `json:"status"`
		Checks  map[string]string `json:"checks"`
		Message string            `json:"message,omitempty"`
	}
}

// healthLive provides a basic liveness check.
// Returns 200 if the server process is running.
func (e *Endpoints) healthLive(ctx context.Context, _ *struct{}) (*HealthLiveResponse, error) {
	return &HealthLiveResponse{
		Body: struct {
			Status string `json:"status"`
		}{
			Status: "ok",
		},
	}, nil
}

// healthReady performs readiness checks for all dependencies.
// Returns 200 if all dependencies are reachable, 503 otherwise.
func (e *Endpoints) healthReady(ctx context.Context, _ *struct{}) (*HealthReadyResponse, error) {
	checks := make(map[string]string)
	allHealthy := true

	// Check database connectivity
	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := e.ravel.Ping(checkCtx); err != nil {
		checks["database"] = "unhealthy: " + err.Error()
		allHealthy = false
	} else {
		checks["database"] = "ok"
	}

	// Future: Add NATS connectivity check
	// Future: Add Containerd connectivity check (for daemon)

	if allHealthy {
		return &HealthReadyResponse{
			Body: struct {
				Status  string            `json:"status"`
				Checks  map[string]string `json:"checks"`
				Message string            `json:"message,omitempty"`
			}{
				Status: "ok",
				Checks: checks,
			},
		}, nil
	}

	return nil, huma.Error503ServiceUnavailable("Service not ready")
}

// RegisterHealthEndpoints adds health check endpoints to the API.
// These should be called separately to avoid authentication requirements.
func (e *Endpoints) RegisterHealthEndpoints(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "healthLive",
		Path:        "/health/live",
		Method:      http.MethodGet,
		Tags:        []string{"health"},
		Summary:     "Liveness check",
		Description: "Returns 200 if the server process is alive. Use for liveness probes.",
	}, e.healthLive)

	huma.Register(api, huma.Operation{
		OperationID: "healthReady",
		Path:        "/health/ready",
		Method:      http.MethodGet,
		Tags:        []string{"health"},
		Summary:     "Readiness check",
		Description: "Returns 200 if all dependencies are reachable, 503 otherwise. Use for readiness probes.",
	}, e.healthReady)
}
