package server

import (
	"context"
	"log/slog"

	"github.com/valyentdev/ravel/core/daemon"
	"github.com/valyentdev/ravel/runtime/disks"
)

type CreateDiskRequest struct {
	Body daemon.DiskOptions
}

type CreateDiskResponse struct {
	Body disks.Disk
}

func (s *DaemonServer) createDisk(ctx context.Context, r *CreateDiskRequest) (*CreateDiskResponse, error) {
	slog.Info("createDisk")
	d, err := s.daemon.CreateDisk(ctx, r.Body)
	if err != nil {
		slog.Error("error creating disk: %v", "err", err)
		return nil, err
	}
	return &CreateDiskResponse{Body: *d}, nil
}

type GetDiskRequest struct {
	Id string `path:"id"`
}

type GetDiskResponse struct {
	Body disks.Disk
}

func (s *DaemonServer) getDisk(ctx context.Context, r *GetDiskRequest) (*GetDiskResponse, error) {
	d, err := s.daemon.GetDisk(ctx, r.Id)
	if err != nil {
		s.log("error getting disk: %v", err)
		return nil, err
	}
	return &GetDiskResponse{Body: *d}, nil
}

type ListDisksRequest struct{}

type ListDisksResponse struct {
	Body []disks.Disk
}

func (s *DaemonServer) listDisks(ctx context.Context, r *ListDisksRequest) (*ListDisksResponse, error) {
	d, err := s.daemon.ListDisks(ctx)
	if err != nil {
		s.log("error listing disks: %v", err)
		return nil, err
	}
	return &ListDisksResponse{Body: d}, nil
}

type DestroyDiskRequest struct {
	Id string `path:"id"`
}

type DestroyDiskResponse struct{}

func (s *DaemonServer) destroyDisk(ctx context.Context, r *DestroyDiskRequest) (*DestroyDiskResponse, error) {
	err := s.daemon.DestroyDisk(ctx, r.Id)
	if err != nil {
		s.log("error destroying disk: %v", err)
		return nil, err
	}
	return &DestroyDiskResponse{}, nil
}
