package agent

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/valyentdev/ravel/internal/server/utils"
	"github.com/valyentdev/ravel/pkg/core"
)

type AgentServer struct {
	server *http.Server
	agent  core.Agent
}

func (e *AgentServer) log(msg string, err error) {
	var rerr *core.RavelError
	if errors.As(err, &rerr) {
		if core.IsUnknown(err) || core.IsInternal(err) {
			slog.Error(msg, "error", err)
		}
	} else {
		slog.Error(msg, "error", err)
	}
}

func (s *AgentServer) ListenAndServe() error {
	return s.server.ListenAndServe()
}

func (s *AgentServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func NewAgentServer(agent core.Agent, address string) *AgentServer {
	as := &AgentServer{agent: agent}

	mux := http.NewServeMux()
	as.registerEndpoints(mux)

	server := &http.Server{
		Addr:    address,
		Handler: mux,
	}

	as.server = server

	return as
}

type CreateInstanceRequest struct {
	Body core.CreateInstancePayload
}

type CreateInstanceResponse struct {
	Body *core.Instance
}

func (s *AgentServer) createInstance(ctx context.Context, req *CreateInstanceRequest) (*CreateInstanceResponse, error) {
	res, err := s.agent.CreateInstance(ctx, req.Body)
	if err != nil {
		s.log("Failed to create instance", err)
		return nil, err
	}
	return &CreateInstanceResponse{Body: res}, nil
}

type ListInstancesRequest struct {
}

type ListInstancesResponse struct {
	Body []core.Instance
}

func (s *AgentServer) listInstances(ctx context.Context, req *ListInstancesRequest) (*ListInstancesResponse, error) {
	res, err := s.agent.ListInstances(ctx)
	if err != nil {
		s.log("Failed to list instances", err)
		return nil, err
	}
	return &ListInstancesResponse{Body: res}, nil
}

type GetInstanceRequest struct {
	Id string `path:"id"`
}

type GetInstanceResponse struct {
	Body core.Instance
}

func (s *AgentServer) getInstance(ctx context.Context, req *GetInstanceRequest) (*GetInstanceResponse, error) {
	res, err := s.agent.GetInstance(ctx, req.Id)
	if err != nil {
		s.log("Failed to get instance", err)
		return nil, err
	}
	return &GetInstanceResponse{Body: *res}, nil
}

type DestroyInstanceRequest struct {
	Id    string `path:"id"`
	Force bool   `query:"force"`
}

type DestroyInstanceResponse struct {
}

func (s *AgentServer) destroyInstance(ctx context.Context, req *DestroyInstanceRequest) (*DestroyInstanceResponse, error) {
	err := s.agent.DestroyInstance(ctx, req.Id, req.Force)
	if err != nil {
		s.log("Failed to destroy instance", err)
		return nil, err
	}
	return &DestroyInstanceResponse{}, nil
}

type StartInstanceRequest struct {
	Id string `path:"id"`
}

type StartInstanceResponse struct {
}

func (s *AgentServer) startInstance(ctx context.Context, req *StartInstanceRequest) (*StartInstanceResponse, error) {
	err := s.agent.StartInstance(ctx, req.Id)
	if err != nil {
		s.log("Failed to start instance", err)
		return nil, err
	}
	return &StartInstanceResponse{}, nil
}

type StopInstanceRequest struct {
	Id   string           `path:"id"`
	Body *core.StopConfig `required:"false"`
}

type StopInstanceResponse struct {
}

func (s *AgentServer) stopInstance(ctx context.Context, req *StopInstanceRequest) (*StopInstanceResponse, error) {
	err := s.agent.StopInstance(ctx, req.Id, req.Body)
	if err != nil {
		s.log("Failed to stop instance", err)
		return nil, err
	}
	return &StopInstanceResponse{}, nil
}

type ExecBody struct {
	Cmd     []string `json:"cmd"`
	Timeout int      `json:"timeout"`
}

type ExecRequest struct {
	Id   string `path:"id"`
	Body core.InstanceExecOptions
}

type ExecResponse struct {
	Body *core.ExecResult
}

func (s *AgentServer) exec(ctx context.Context, req *ExecRequest) (*ExecResponse, error) {
	res, err := s.agent.InstanceExec(ctx, req.Id, req.Body)
	if err != nil {
		s.log("Failed to exec command", err)
		return nil, err
	}
	return &ExecResponse{Body: res}, nil
}

type GetInstanceLogsRequest struct {
	Id     string `path:"id"`
	Follow bool   `query:"follow"`
}

type GetInstanceLogsResponse struct {
	Body []byte
}

func (s *AgentServer) getInstanceLogs(ctx context.Context, req *GetInstanceLogsRequest) (*huma.StreamResponse, error) {
	var err error
	var logsChan <-chan []*core.LogEntry
	var logs []*core.LogEntry
	if req.Follow {
		logsChan, err = s.agent.SubscribeToInstanceLogs(ctx, req.Id)
	} else {
		logs, err = s.agent.GetInstanceLogs(ctx, req.Id)
	}

	if err != nil {
		s.log("Failed to get instance logs", err)
		return nil, err
	}

	slog.Info("Getting instance logs", "id", req.Id, "follow", req.Follow)

	return &huma.StreamResponse{
		Body: func(ctx huma.Context) {
			ctx.SetHeader("Content-Type", "application/x-ndjson")
			ctx.SetStatus(http.StatusOK)

			if req.Follow {
				streamLogs(ctx, logsChan)
				return
			}

			bw := ctx.BodyWriter()
			rw := bw.(http.ResponseWriter)
			rc := http.NewResponseController(rw)

			err := rc.SetWriteDeadline(time.Now().Add(3 * time.Second))

			if err != nil {
				return
			}

			for _, entry := range logs {
				slog.Info("Writing log entry", "entry", entry)
				bytes, err := json.Marshal(entry)
				if err != nil {
					return
				}

				_, err = bw.Write(bytes)
				if err != nil {
					return
				}
				_, err = bw.Write([]byte("\n"))

				if err != nil {
					return
				}
			}

		},
	}, nil
}

func streamLogs(ctx huma.Context, logsChan <-chan []*core.LogEntry) {
	bw := ctx.BodyWriter()
	rw := bw.(http.ResponseWriter)
	rc := http.NewResponseController(rw)

	err := rc.SetWriteDeadline(time.Now().Add(30 * time.Second))
	if err != nil {
		slog.Error("Failed to set write deadline", "error", err)
		return
	}
	err = rc.Flush()
	if err != nil {
		slog.Error("Failed to flush response", "error", err)
		return
	}

	for log := range logsChan {
		for _, entry := range log {
			bytes, err := json.Marshal(entry)
			if err != nil {
				return
			}

			_, err = rw.Write(bytes)
			if err != nil {
				return
			}
			_, err = rw.Write([]byte("\n"))

			if err != nil {
				return
			}

			err = rc.Flush()
			if err != nil {
				return
			}

		}
	}
}

func (s AgentServer) registerEndpoints(mux humago.Mux) {
	humaConfig := utils.GetHumaConfig()
	api := humago.New(mux, humaConfig)

	huma.Register(api, huma.Operation{
		OperationID: "createInstance",
		Path:        "/instances",
		Method:      http.MethodPost,
	}, s.createInstance)

	huma.Register(api, huma.Operation{
		OperationID: "listInstances",
		Path:        "/instances",
		Method:      http.MethodGet,
	}, s.listInstances)

	huma.Register(api, huma.Operation{
		OperationID: "getInstance",
		Path:        "/instances/{id}",
		Method:      http.MethodGet,
	}, s.getInstance)

	huma.Register(api, huma.Operation{
		OperationID: "destroyInstance",
		Path:        "/instances/{id}",
		Method:      http.MethodDelete,
	}, s.destroyInstance)

	huma.Register(api, huma.Operation{
		OperationID: "startInstance",
		Path:        "/instances/{id}/start",
		Method:      http.MethodPost,
	}, s.startInstance)

	huma.Register(api, huma.Operation{
		OperationID: "stopInstance",
		Path:        "/instances/{id}/stop",
		Method:      http.MethodPost,
	}, s.stopInstance)

	huma.Register(api, huma.Operation{
		OperationID: "exec",
		Path:        "/instances/{id}/exec",
		Method:      http.MethodPost,
	}, s.exec)

	huma.Register(api, huma.Operation{
		OperationID: "getInstanceLogs",
		Path:        "/instances/{id}/logs",
		Method:      http.MethodGet,
	}, s.getInstanceLogs)

}
