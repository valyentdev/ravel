package server

import (
	"context"
	"time"

	"github.com/alexisbouchez/ravel/api"
	"github.com/alexisbouchez/ravel/core/daemon"
	"github.com/alexisbouchez/ravel/core/instance"
	"github.com/alexisbouchez/ravel/internal/streamutil"
	"github.com/danielgtaylor/huma/v2"
)

type CreateInstanceRequest struct {
	Body daemon.InstanceOptions
}

type CreateInstanceResponse struct {
	Body *instance.Instance
}

func (s *DaemonServer) createInstance(ctx context.Context, req *CreateInstanceRequest) (*CreateInstanceResponse, error) {
	res, err := s.daemon.CreateInstance(ctx, req.Body)
	if err != nil {
		s.log("Failed to create instance", err)
		return nil, err
	}
	return &CreateInstanceResponse{Body: res}, nil
}

type ListInstancesRequest struct {
}

type ListInstancesResponse struct {
	Body []instance.Instance
}

func (s *DaemonServer) listInstances(ctx context.Context, req *ListInstancesRequest) (*ListInstancesResponse, error) {
	res, err := s.daemon.ListInstances(ctx)
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
	Body instance.Instance
}

func (s *DaemonServer) getInstance(ctx context.Context, req *GetInstanceRequest) (*GetInstanceResponse, error) {
	res, err := s.daemon.GetInstance(ctx, req.Id)
	if err != nil {
		s.log("Failed to get instance", err)
		return nil, err
	}
	return &GetInstanceResponse{Body: *res}, nil
}

type DestroyInstanceRequest struct {
	Id string `path:"id"`
}

type DestroyInstanceResponse struct {
}

func (s *DaemonServer) destroyInstance(ctx context.Context, req *DestroyInstanceRequest) (*DestroyInstanceResponse, error) {
	err := s.daemon.DestroyInstance(ctx, req.Id)
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

func (s *DaemonServer) startInstance(ctx context.Context, req *StartInstanceRequest) (*StartInstanceResponse, error) {
	err := s.daemon.StartInstance(ctx, req.Id)
	if err != nil {
		s.log("Failed to start instance", err)
		return nil, err
	}
	return &StartInstanceResponse{}, nil
}

// StartInstanceFromSnapshotRequest is the request for starting an instance from a snapshot
type StartInstanceFromSnapshotRequest struct {
	Id   string `path:"id"`
	Body struct {
		SnapshotId string `json:"snapshot_id" required:"true" doc:"Snapshot ID to restore from"`
	}
}

type StartInstanceFromSnapshotResponse struct {
}

func (s *DaemonServer) startInstanceFromSnapshot(ctx context.Context, req *StartInstanceFromSnapshotRequest) (*StartInstanceFromSnapshotResponse, error) {
	// Global snapshot path and jail-relative path
	globalSnapshotPath := "/var/lib/ravel/global-snapshots/" + req.Id + "/" + req.Body.SnapshotId
	jailSnapshotPath := "/snapshots/" + req.Body.SnapshotId
	err := s.daemon.StartInstanceFromSnapshot(ctx, req.Id, globalSnapshotPath, jailSnapshotPath)
	if err != nil {
		s.log("Failed to start instance from snapshot", err)
		return nil, err
	}
	return &StartInstanceFromSnapshotResponse{}, nil
}

type StopInstanceRequest struct {
	Id   string          `path:"id"`
	Body *api.StopConfig `required:"false"`
}

type StopInstanceResponse struct {
}

func (s *DaemonServer) stopInstance(ctx context.Context, req *StopInstanceRequest) (*StopInstanceResponse, error) {
	err := s.daemon.StopInstance(ctx, req.Id, req.Body)
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
	Body api.ExecOptions
}

type ExecResponse struct {
	Body *api.ExecResult
}

func (s *DaemonServer) exec(ctx context.Context, req *ExecRequest) (*ExecResponse, error) {
	res, err := s.daemon.InstanceExec(ctx, req.Id, req.Body.Cmd, time.Duration(req.Body.TimeoutMs)*time.Millisecond)
	if err != nil {
		s.log("Failed to exec command", err)
		return nil, err
	}
	return &ExecResponse{Body: res}, nil
}

type FollowInstanceLogsRequest struct {
	Id string `path:"id"`
}

func (s *DaemonServer) followInstanceLogs(ctx context.Context, req *GetInstanceLogsRequest) (*huma.StreamResponse, error) {
	logs, logsChan, err := s.daemon.SubscribeToInstanceLogs(ctx, req.Id)
	if err != nil {
		s.log("Failed to get instance logs", err)
		return nil, err
	}

	return &huma.StreamResponse{
		Body: func(ctx huma.Context) {
			streamutil.StreamLogs(ctx, logs, logsChan)
		}}, nil
}

type GetInstanceLogsRequest struct {
	Id     string `path:"id"`
	Follow bool   `query:"follow"`
}

type GetInstanceLogsResponse struct {
	Body []*api.LogEntry
}

func (s *DaemonServer) getInstanceLogs(ctx context.Context, req *GetInstanceLogsRequest) (*GetInstanceLogsResponse, error) {
	logs, err := s.daemon.GetInstanceLogs(ctx, req.Id)
	if err != nil {
		s.log("Failed to get instance logs", err)
		return nil, err
	}

	return &GetInstanceLogsResponse{Body: logs}, nil
}

// Snapshot API for AI sandbox fast starts

type InstanceSnapshotRequest struct {
	Id   string `path:"id"`
	Body struct {
		SnapshotId string `json:"snapshot_id" required:"true" doc:"Unique identifier for the snapshot"`
	}
}

type InstanceSnapshotResponse struct {
	Body struct {
		SnapshotId string `json:"snapshot_id"`
		Path       string `json:"path"`
	}
}

func (s *DaemonServer) instanceSnapshot(ctx context.Context, req *InstanceSnapshotRequest) (*InstanceSnapshotResponse, error) {
	// Snapshot path relative to jail chroot - CloudHypervisor runs inside the jail
	// so /snapshots becomes /var/lib/ravel/instances/{id}/snapshots on host
	jailRelativePath := "/snapshots/" + req.Body.SnapshotId
	err := s.daemon.InstanceSnapshot(ctx, req.Id, jailRelativePath)
	if err != nil {
		s.log("Failed to snapshot instance", err)
		return nil, err
	}

	return &InstanceSnapshotResponse{
		Body: struct {
			SnapshotId string `json:"snapshot_id"`
			Path       string `json:"path"`
		}{
			SnapshotId: req.Body.SnapshotId,
			Path:       jailRelativePath,
		},
	}, nil
}

type InstanceRestoreRequest struct {
	Id   string `path:"id"`
	Body struct {
		SnapshotId string `json:"snapshot_id" required:"true" doc:"Snapshot ID to restore from"`
	}
}

type InstanceRestoreResponse struct {
}

func (s *DaemonServer) instanceRestore(ctx context.Context, req *InstanceRestoreRequest) (*InstanceRestoreResponse, error) {
	// Snapshot path relative to jail chroot
	jailRelativePath := "/snapshots/" + req.Body.SnapshotId
	err := s.daemon.InstanceRestore(ctx, req.Id, jailRelativePath)
	if err != nil {
		s.log("Failed to restore instance", err)
		return nil, err
	}

	return &InstanceRestoreResponse{}, nil
}
