package server

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"os"

	"github.com/alexisbouchez/ravel/agent/build"
	"github.com/alexisbouchez/ravel/api"
	"github.com/danielgtaylor/huma/v2"
)

// BuildService interface for build operations
type BuildService interface {
	IsEnabled() bool
	StartBuild(ctx context.Context, opts build.BuildOptions) (*api.Build, error)
	GetBuild(ctx context.Context, id string) (*api.Build, error)
	ListBuilds(ctx context.Context, namespace string, limit int) ([]*api.Build, error)
	GetBuildLogs(buildId string) (string, error)
	CancelBuild(id string) error
}

// SetBuildService sets the build service on the agent server
func (s *AgentServer) SetBuildService(bs BuildService) {
	s.buildService = bs
}

// CreateBuildRequest is the request for creating a new build
type CreateBuildRequest struct {
	Namespace  string `query:"namespace" required:"true"`
	ImageName  string `query:"image_name" required:"true"`
	Tag        string `query:"tag"`
	Registry   string `query:"registry" required:"true"`
	Dockerfile string `query:"dockerfile"`
	Target     string `query:"target"`
	NoCache    bool   `query:"no_cache"`
	RawBody    huma.MultipartFormFiles[struct {
		Context huma.FormFile `form:"context" required:"true"`
	}]
}

type CreateBuildResponse struct {
	Body *api.Build
}

func (s *AgentServer) createBuild(ctx context.Context, req *CreateBuildRequest) (*CreateBuildResponse, error) {
	if s.buildService == nil || !s.buildService.IsEnabled() {
		return nil, huma.Error503ServiceUnavailable("build service is not enabled")
	}

	data := req.RawBody.Data()
	contextFile := data.Context

	if !contextFile.IsSet {
		return nil, huma.Error400BadRequest("context file is required")
	}

	opts := build.BuildOptions{
		Namespace:  req.Namespace,
		ImageName:  req.ImageName,
		Tag:        req.Tag,
		Registry:   req.Registry,
		Dockerfile: req.Dockerfile,
		Target:     req.Target,
		NoCache:    req.NoCache,
		Context:    contextFile.File,
	}

	b, err := s.buildService.StartBuild(ctx, opts)
	if err != nil {
		s.log("Failed to start build", err)
		return nil, err
	}

	return &CreateBuildResponse{Body: b}, nil
}

type GetBuildRequest struct {
	Id string `path:"id"`
}

type GetBuildResponse struct {
	Body *api.Build
}

func (s *AgentServer) getBuild(ctx context.Context, req *GetBuildRequest) (*GetBuildResponse, error) {
	if s.buildService == nil {
		return nil, huma.Error503ServiceUnavailable("build service is not enabled")
	}

	b, err := s.buildService.GetBuild(ctx, req.Id)
	if err != nil {
		s.log("Failed to get build", err)
		return nil, err
	}

	return &GetBuildResponse{Body: b}, nil
}

type ListBuildsRequest struct {
	Namespace string `query:"namespace"`
	Limit     int    `query:"limit"`
}

type ListBuildsResponse struct {
	Body []*api.Build
}

func (s *AgentServer) listBuilds(ctx context.Context, req *ListBuildsRequest) (*ListBuildsResponse, error) {
	if s.buildService == nil {
		return nil, huma.Error503ServiceUnavailable("build service is not enabled")
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}

	builds, err := s.buildService.ListBuilds(ctx, req.Namespace, limit)
	if err != nil {
		s.log("Failed to list builds", err)
		return nil, err
	}

	return &ListBuildsResponse{Body: builds}, nil
}

type GetBuildLogsRequest struct {
	Id     string `path:"id"`
	Follow bool   `query:"follow"`
}

func (s *AgentServer) getBuildLogs(ctx context.Context, req *GetBuildLogsRequest) (*huma.StreamResponse, error) {
	if s.buildService == nil {
		return nil, huma.Error503ServiceUnavailable("build service is not enabled")
	}

	logPath, err := s.buildService.GetBuildLogs(req.Id)
	if err != nil {
		return nil, huma.Error404NotFound("build logs not found")
	}

	return &huma.StreamResponse{
		Body: func(ctx huma.Context) {
			ctx.SetHeader("Content-Type", "text/plain")
			ctx.SetStatus(http.StatusOK)

			file, err := os.Open(logPath)
			if err != nil {
				return
			}
			defer file.Close()

			bw := ctx.BodyWriter()
			rw, ok := bw.(http.ResponseWriter)
			if !ok {
				io.Copy(bw, file)
				return
			}

			rc := http.NewResponseController(rw)
			reader := bufio.NewReader(file)

			for {
				line, err := reader.ReadBytes('\n')
				if len(line) > 0 {
					rw.Write(line)
					rc.Flush()
				}
				if err == io.EOF {
					if !req.Follow {
						break
					}
					// For follow mode, wait and try again
					select {
					case <-ctx.Context().Done():
						return
					default:
						continue
					}
				}
				if err != nil {
					break
				}
			}
		},
	}, nil
}

type CancelBuildRequest struct {
	Id string `path:"id"`
}

type CancelBuildResponse struct{}

func (s *AgentServer) cancelBuild(ctx context.Context, req *CancelBuildRequest) (*CancelBuildResponse, error) {
	if s.buildService == nil {
		return nil, huma.Error503ServiceUnavailable("build service is not enabled")
	}

	if err := s.buildService.CancelBuild(req.Id); err != nil {
		s.log("Failed to cancel build", err)
		return nil, err
	}

	return &CancelBuildResponse{}, nil
}
