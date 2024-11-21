package server

import (
	"context"

	"github.com/valyentdev/ravel/runtime"
)

type ListImagesRequest struct {
}

type ListImagesResponse struct {
	Body []runtime.Image
}

func (s *AgentServer) listImages(ctx context.Context, req *ListImagesRequest) (*ListImagesResponse, error) {
	res, err := s.agent.ListImages(ctx)
	if err != nil {
		s.log("Failed to list images", err)
		return nil, err
	}
	return &ListImagesResponse{Body: res}, nil
}

type PullImageRequest struct {
	Body runtime.PullImageOptions
}

type PullImageResponse struct {
	Body *runtime.Image
}

func (s *AgentServer) pullImage(ctx context.Context, req *PullImageRequest) (*PullImageResponse, error) {
	res, err := s.agent.PullImage(ctx, req.Body)
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

func (s *AgentServer) deleteImage(ctx context.Context, req *DeleteImageRequest) (*DeleteImageResponse, error) {
	err := s.agent.DeleteImage(ctx, req.Ref)
	if err != nil {
		s.log("Failed to delete image", err)
		return nil, err
	}
	return &DeleteImageResponse{}, nil
}
