package endpoints

import (
	"bufio"
	"context"
	"io"
	"net/http"

	"github.com/alexisbouchez/ravel/api"
	"github.com/alexisbouchez/ravel/ravel"
	"github.com/danielgtaylor/huma/v2"
)

// CreateBuildRequest is the request for creating a new build
type CreateBuildRequest struct {
	Namespace  string `path:"namespace"`
	ImageName  string `query:"image_name" required:"true" doc:"Target image name"`
	Tag        string `query:"tag" doc:"Image tag (default: latest)"`
	Registry   string `query:"registry" required:"true" doc:"Registry URL to push to"`
	Dockerfile string `query:"dockerfile" doc:"Path to Dockerfile in context"`
	Target     string `query:"target" doc:"Target build stage"`
	NoCache    bool   `query:"no_cache" doc:"Disable build cache"`
	RawBody    huma.MultipartFormFiles[struct {
		Context huma.FormFile `form:"context" required:"true"`
	}]
}

type CreateBuildResponse struct {
	Body *api.Build
}

func (e *Endpoints) createBuild(ctx context.Context, req *CreateBuildRequest) (*CreateBuildResponse, error) {
	data := req.RawBody.Data()

	if !data.Context.IsSet {
		return nil, huma.Error400BadRequest("context file is required")
	}

	build, err := e.ravel.CreateBuild(ctx, ravel.CreateBuildOptions{
		Namespace:  req.Namespace,
		ImageName:  req.ImageName,
		Tag:        req.Tag,
		Registry:   req.Registry,
		Dockerfile: req.Dockerfile,
		Target:     req.Target,
		NoCache:    req.NoCache,
		Context:    data.Context.File,
	})
	if err != nil {
		e.log("Failed to create build", err)
		return nil, err
	}

	return &CreateBuildResponse{Body: build}, nil
}

type GetBuildRequest struct {
	Namespace string `path:"namespace"`
	BuildId   string `path:"build_id"`
}

type GetBuildResponse struct {
	Body *api.Build
}

func (e *Endpoints) getBuild(ctx context.Context, req *GetBuildRequest) (*GetBuildResponse, error) {
	build, err := e.ravel.GetBuild(ctx, req.Namespace, req.BuildId)
	if err != nil {
		e.log("Failed to get build", err)
		return nil, err
	}

	return &GetBuildResponse{Body: build}, nil
}

type ListBuildsRequest struct {
	Namespace string `path:"namespace"`
	Limit     int    `query:"limit"`
}

type ListBuildsResponse struct {
	Body []*api.Build
}

func (e *Endpoints) listBuilds(ctx context.Context, req *ListBuildsRequest) (*ListBuildsResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}

	builds, err := e.ravel.ListBuilds(ctx, req.Namespace, limit)
	if err != nil {
		e.log("Failed to list builds", err)
		return nil, err
	}

	return &ListBuildsResponse{Body: builds}, nil
}

type GetBuildLogsRequest struct {
	Namespace string `path:"namespace"`
	BuildId   string `path:"build_id"`
	Follow    bool   `query:"follow"`
}

func (e *Endpoints) getBuildLogs(ctx context.Context, req *GetBuildLogsRequest) (*huma.StreamResponse, error) {
	logs, err := e.ravel.GetBuildLogs(ctx, req.Namespace, req.BuildId, req.Follow)
	if err != nil {
		return nil, huma.Error404NotFound("build logs not found")
	}

	return &huma.StreamResponse{
		Body: func(ctx huma.Context) {
			defer logs.Close()

			ctx.SetHeader("Content-Type", "text/plain")
			ctx.SetStatus(http.StatusOK)

			bw := ctx.BodyWriter()
			rw, ok := bw.(http.ResponseWriter)
			if !ok {
				io.Copy(bw, logs)
				return
			}

			rc := http.NewResponseController(rw)
			reader := bufio.NewReader(logs)

			for {
				line, err := reader.ReadBytes('\n')
				if len(line) > 0 {
					rw.Write(line)
					rc.Flush()
				}
				if err == io.EOF {
					break
				}
				if err != nil {
					break
				}
			}
		},
	}, nil
}

type CancelBuildRequest struct {
	Namespace string `path:"namespace"`
	BuildId   string `path:"build_id"`
}

type CancelBuildResponse struct{}

func (e *Endpoints) cancelBuild(ctx context.Context, req *CancelBuildRequest) (*CancelBuildResponse, error) {
	if err := e.ravel.CancelBuild(ctx, req.Namespace, req.BuildId); err != nil {
		e.log("Failed to cancel build", err)
		return nil, err
	}

	return &CancelBuildResponse{}, nil
}
