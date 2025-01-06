package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/api/initd"
	"github.com/valyentdev/ravel/core/errdefs"
	"github.com/valyentdev/ravel/initd/exec"
	"github.com/valyentdev/ravel/initd/files"
)

func (e *publicEndpoints) registerRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		Method:      "GET",
		Path:        "/status",
		OperationID: "getStatus",
		Summary:     "Get status of Initd",
	}, e.getStatus)

	huma.Register(api, huma.Operation{
		Method:      "POST",
		Path:        "/exec",
		OperationID: "exec",
		Description: "Execute a command",
	}, e.exec)

	huma.Register(api, huma.Operation{
		Method:      "GET",
		Path:        "/fs/ls",
		OperationID: "listDir",
		Summary:     "List directory contents",
	}, e.listDir)

	huma.Register(api, huma.Operation{
		Method:      "POST",
		Path:        "/fs/mkdir",
		OperationID: "makeDirectory",
		Summary:     "Create a directory",
	}, e.mkdir)

	huma.Register(api, huma.Operation{
		Method:      "GET",
		Path:        "/fs/read",
		OperationID: "readFile",
		Summary:     "Read a file",
	}, e.readFile)

	huma.Register(api, huma.Operation{
		Method:      "POST",
		Path:        "/fs/rm",
		OperationID: "remove",
		Summary:     "Remove a file/directory",
	}, e.remove)

	huma.Register(api, huma.Operation{
		Method:           "POST",
		Path:             "/fs/write",
		OperationID:      "writeFile",
		Summary:          "Upload a file",
		SkipValidateBody: true,
		RequestBody: &huma.RequestBody{
			Content: map[string]*huma.MediaType{
				"multipart/form-data": {
					Schema: &huma.Schema{
						Type: "object",
						Properties: map[string]*huma.Schema{
							"path": {
								Type: "string",
							},
							"file": {
								Type:   "string",
								Format: "binary",
							},
						},
					},
				},
			},
		},
	}, e.writeFile)

	huma.Register(api, huma.Operation{
		Method:      "GET",
		Path:        "/fs/watch",
		OperationID: "watchDir",
		Summary:     "Watch a directory",
	}, e.watchDir)

	huma.Register(api, huma.Operation{
		Method:      "GET",
		Path:        "/fs/stat",
		OperationID: "statFile",
		Summary:     "Get file infos",
	}, e.statFile)

}

type publicEndpoints struct {
	files *files.Service
}

type ListDirRequest struct {
	Path string `query:"path"`
}

type ListDirResponse struct {
	Body []initd.FSEntry
}

func (e *publicEndpoints) listDir(ctx context.Context, req *ListDirRequest) (*ListDirResponse, error) {
	entries, err := e.files.ListDir(ctx, req.Path)
	if err != nil {
		return nil, err
	}
	return &ListDirResponse{Body: entries}, nil
}

type MkdirRequest struct {
	Body initd.MkdirOptions
}

type MkdirResponse struct {
}

func (e *publicEndpoints) mkdir(ctx context.Context, req *MkdirRequest) (*MkdirResponse, error) {
	err := e.files.Mkdir(ctx, req.Body.Dir)
	if err != nil {
		return nil, err
	}
	return &MkdirResponse{}, nil
}

type ReadFileRequest struct {
	Path string `query:"path"`
}

func (e *publicEndpoints) readFile(ctx context.Context, req *ReadFileRequest) (*huma.StreamResponse, error) {
	file, err := e.files.OpenFile(ctx, req.Path)
	if err != nil {
		return nil, err
	}

	name := path.Base(req.Path)

	return &huma.StreamResponse{
		Body: func(ctx huma.Context) {
			defer file.Close()
			ctx.SetHeader("Content-Type", "application/octet-stream")
			ctx.SetHeader("Content-Disposition", fmt.Sprintf("attachment; filename=%s", name))
			writer := ctx.BodyWriter()

			_, err := io.Copy(writer, file)
			if err != nil {
				ctx.SetStatus(500)
			}
		},
	}, nil
}

type RemoveRequest struct {
	Path string `query:"path"`
}

type RemoveResponse struct {
}

func (e *publicEndpoints) remove(ctx context.Context, req *RemoveRequest) (*RemoveResponse, error) {
	err := e.files.Remove(ctx, req.Path)
	if err != nil {
		return nil, err
	}
	return &RemoveResponse{}, nil
}

type ExecRequest struct {
	Body api.ExecOptions
}

type ExecResponse struct {
	Body *api.ExecResult
}

func (e *publicEndpoints) exec(ctx context.Context, req *ExecRequest) (*ExecResponse, error) {
	result, err := exec.Exec(ctx, req.Body)
	if err != nil {
		return nil, err
	}
	return &ExecResponse{Body: result}, nil
}

type GetStatusRequest struct {
}

type GetStatusResponse struct {
	Body initd.Status
}

func (e *publicEndpoints) getStatus(ctx context.Context, req *GetStatusRequest) (*GetStatusResponse, error) {
	return &GetStatusResponse{Body: initd.Status{Ok: true}}, nil
}

type WriteFileRequest struct {
	RawBody multipart.Form
}

type WriteFileResponse struct {
}

func (e *publicEndpoints) writeFile(ctx context.Context, req *WriteFileRequest) (*WriteFileResponse, error) {
	form := req.RawBody

	pathParts, ok := form.Value["path"]
	if !ok {
		return nil, errdefs.NewInvalidArgument("missing a path")
	}

	if len(pathParts) != 1 {
		return nil, errdefs.NewInvalidArgument("only one path is allowed")
	}

	p := pathParts[0]

	filePart, ok := form.File["file"]
	if !ok {
		return nil, errdefs.NewInvalidArgument("missing a file")
	}

	if len(filePart) != 1 {
		return nil, errdefs.NewInvalidArgument("only one file is allowed")
	}

	file, err := filePart[0].Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	err = e.files.WriteFile(ctx, p, file)
	if err != nil {
		return nil, err
	}

	return &WriteFileResponse{}, nil
}

type StatFileRequest struct {
	Path string `query:"path"`
}

type StatFileResponse struct {
	Body *initd.FSEntry
}

func (e *publicEndpoints) statFile(ctx context.Context, req *StatFileRequest) (*StatFileResponse, error) {
	entry, err := e.files.Stat(ctx, req.Path)
	if err != nil {
		return nil, err
	}
	return &StatFileResponse{Body: entry}, nil
}

type WatchDirRequest struct {
	Path string `query:"path"`
}

func (e *publicEndpoints) watchDir(ctx context.Context, req *WatchDirRequest) (*huma.StreamResponse, error) {
	events, err := e.files.WatchDir(ctx, req.Path)
	if err != nil {
		return nil, err
	}

	return &huma.StreamResponse{
		Body: func(ctx huma.Context) {
			ctx.SetHeader("Content-Type", "application/x-ndjson")
			bw := ctx.BodyWriter()
			rw := bw.(http.ResponseWriter)
			rc := http.NewResponseController(rw)

			rc.SetWriteDeadline(time.Time{}) // Disable write deadline
			for event := range events {
				bytes, err := json.Marshal(event)
				if err != nil {
					ctx.SetStatus(500)
					return
				}

				_, err = rw.Write(bytes)
				if err != nil {
					ctx.SetStatus(500)
					return
				}

				_, err = rw.Write([]byte("\n"))
				if err != nil {
					ctx.SetStatus(500)
					return
				}

				err = rc.Flush()
				if err != nil {
					ctx.SetStatus(500)
					return
				}

			}
		},
	}, nil
}
