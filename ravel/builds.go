package ravel

import (
	"context"
	"fmt"
	"io"

	agentclient "github.com/alexisbouchez/ravel/agent/client"
	"github.com/alexisbouchez/ravel/api"
)

// CreateBuildOptions contains options for creating a new build
type CreateBuildOptions struct {
	Namespace  string
	ImageName  string
	Tag        string
	Registry   string
	Dockerfile string
	Target     string
	NoCache    bool
	Context    io.Reader // tar.gz stream
}

// CreateBuild starts a new image build
func (r *Ravel) CreateBuild(ctx context.Context, opts CreateBuildOptions) (*api.Build, error) {
	// Validate namespace exists
	_, err := r.GetNamespace(ctx, opts.Namespace)
	if err != nil {
		return nil, err
	}

	// Select a node to run the build on
	// For now, use round-robin selection from available nodes
	nodes, err := r.o.ListNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no available nodes for build")
	}

	// Select the first available node (can be improved with load balancing)
	node := nodes[0]

	// Get agent client for the selected node
	agentClient, err := r.o.GetAgentClient(node.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent client: %w", err)
	}

	// Start build on the agent
	build, err := agentClient.CreateBuild(ctx, agentclient.CreateBuildOptions{
		Namespace:  opts.Namespace,
		ImageName:  opts.ImageName,
		Tag:        opts.Tag,
		Registry:   opts.Registry,
		Dockerfile: opts.Dockerfile,
		Target:     opts.Target,
		NoCache:    opts.NoCache,
		Context:    opts.Context,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start build on agent: %w", err)
	}

	// Add node ID to build info
	build.NodeId = node.Id

	return build, nil
}

// GetBuild gets the status of a build
func (r *Ravel) GetBuild(ctx context.Context, namespace, buildId string) (*api.Build, error) {
	// Validate namespace exists
	_, err := r.GetNamespace(ctx, namespace)
	if err != nil {
		return nil, err
	}

	// For now, we need to query all nodes to find the build
	// This can be optimized by storing build metadata in the database
	nodes, err := r.o.ListNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	for _, node := range nodes {
		agentClient, err := r.o.GetAgentClient(node.Id)
		if err != nil {
			continue
		}

		build, err := agentClient.GetBuild(ctx, buildId)
		if err == nil {
			build.NodeId = node.Id
			return build, nil
		}
	}

	return nil, fmt.Errorf("build not found: %s", buildId)
}

// ListBuilds lists builds in a namespace
func (r *Ravel) ListBuilds(ctx context.Context, namespace string, limit int) ([]*api.Build, error) {
	// Validate namespace exists
	_, err := r.GetNamespace(ctx, namespace)
	if err != nil {
		return nil, err
	}

	// Query all nodes and aggregate builds
	nodes, err := r.o.ListNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	var allBuilds []*api.Build
	for _, node := range nodes {
		agentClient, err := r.o.GetAgentClient(node.Id)
		if err != nil {
			continue
		}

		builds, err := agentClient.ListBuilds(ctx, namespace, limit)
		if err != nil {
			continue
		}

		for _, b := range builds {
			b.NodeId = node.Id
			allBuilds = append(allBuilds, b)
		}
	}

	// Apply limit
	if limit > 0 && len(allBuilds) > limit {
		allBuilds = allBuilds[:limit]
	}

	return allBuilds, nil
}

// GetBuildLogs streams build logs
func (r *Ravel) GetBuildLogs(ctx context.Context, namespace, buildId string, follow bool) (io.ReadCloser, error) {
	// Validate namespace exists
	_, err := r.GetNamespace(ctx, namespace)
	if err != nil {
		return nil, err
	}

	// Find the node with this build
	nodes, err := r.o.ListNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	for _, node := range nodes {
		agentClient, err := r.o.GetAgentClient(node.Id)
		if err != nil {
			continue
		}

		// Try to get logs from this node
		logs, err := agentClient.GetBuildLogs(ctx, buildId, follow)
		if err == nil {
			return logs, nil
		}
	}

	return nil, fmt.Errorf("build not found: %s", buildId)
}

// CancelBuild cancels an in-progress build
func (r *Ravel) CancelBuild(ctx context.Context, namespace, buildId string) error {
	// Validate namespace exists
	_, err := r.GetNamespace(ctx, namespace)
	if err != nil {
		return err
	}

	// Find the node with this build
	nodes, err := r.o.ListNodes(ctx)
	if err != nil {
		return fmt.Errorf("failed to list nodes: %w", err)
	}

	for _, node := range nodes {
		agentClient, err := r.o.GetAgentClient(node.Id)
		if err != nil {
			continue
		}

		// Try to cancel on this node
		err = agentClient.CancelBuild(ctx, buildId)
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("build not found: %s", buildId)
}
