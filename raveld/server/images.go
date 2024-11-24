package server

import (
	"context"
	"net/url"

	"github.com/valyentdev/ravel/core/daemon"
)

type ListImagesRequest struct {
}

type ListImagesResponse struct {
	Body []daemon.Image
}

func (s *DaemonServer) listImages(ctx context.Context, req *ListImagesRequest) (*ListImagesResponse, error) {
	res, err := s.daemon.ListImages(ctx)
	if err != nil {
		s.log("Failed to list images", err)
		return nil, err
	}
	return &ListImagesResponse{Body: res}, nil
}

type PullImageRequest struct {
	Body daemon.ImagePullOptions
}

type PullImageResponse struct {
	Body *daemon.Image
}

func (s *DaemonServer) pullImage(ctx context.Context, req *PullImageRequest) (*PullImageResponse, error) {
	res, err := s.daemon.PullImage(ctx, req.Body)
	if err != nil {
		s.log("Failed to pull image", err)
		return nil, err
	}
	return &PullImageResponse{Body: res}, nil
}

type DeleteImageRequest struct {
	Ref string `path:"ref"`
}

type DeleteImageResponse struct {
}

func (s *DaemonServer) deleteImage(ctx context.Context, req *DeleteImageRequest) (*DeleteImageResponse, error) {
	ref, err := url.PathUnescape(req.Ref)
	if err != nil {
		return nil, err
	}
	err = s.daemon.DeleteImage(ctx, ref)
	if err != nil {
		s.log("Failed to delete image", err)
		return nil, err
	}
	return &DeleteImageResponse{}, nil
}
