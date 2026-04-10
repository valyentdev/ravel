package vm

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/core/snapshots"
	"github.com/containerd/errdefs"
	"github.com/opencontainers/image-spec/identity"
)

func (b *Driver) removeRootFS(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := b.snapshotter.Remove(ctx, rootFSName(id))
	if err != nil {
		if errdefs.IsNotFound(err) {
			return nil
		}

		return fmt.Errorf("failed to remove snapshot %q: %w", id, err)
	}

	return nil
}

func (b *Driver) prepareRootFS(ctx context.Context, id string, image client.Image) (rootfs string, err error) {
	rootfs, err = b.prepareContainerRootFS(ctx, id, image)
	if err != nil {
		return "", fmt.Errorf("failed to prepare rootfs for instance %q: %w", id, err)
	}

	return rootfs, nil
}

func rootFSName(id string) string {
	return fmt.Sprintf("%s-%s", id, "rootfs")
}

func (b *Driver) prepareContainerRootFS(ctx context.Context, id string, image client.Image) (string, error) {
	slog.Debug("preparing rootfs for container", "id", id)
	diffIDs, err := image.RootFS(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get rootfs for image %q: %w", image.Name(), err)
	}

	parent := identity.ChainID(diffIDs).String()

	slog.Debug("preparing snapshot", "id", id, "parent", parent)

	labels := map[string]string{
		"containerd.io/gc.root": time.Now().UTC().Format(time.RFC3339),
	}

	mounts, err := b.snapshotter.Prepare(ctx, rootFSName(id), parent, snapshots.WithLabels(labels))
	if err != nil {
		if !errdefs.IsAlreadyExists(err) {
			return "", fmt.Errorf("failed to prepare snapshot %q: %w", id, err)
		}

		slog.Debug("snapshot already exists, removing", "id", id)
		err = b.removeRootFS(id)
		if err != nil {
			return "", fmt.Errorf("failed to remove existing snapshot %q: %w", id, err)
		}

		slog.Debug("retrying snapshot preparation", "id", id, "parent", parent)
		mounts, err = b.snapshotter.Prepare(context.Background(), rootFSName(id), parent, snapshots.WithLabels(labels))
		if err != nil {
			return "", fmt.Errorf("failed to prepare snapshot %q: %w", id, err)
		}
	}

	if len(mounts) == 0 {
		return "", fmt.Errorf("no mounts found for container %q", id)
	}

	return mounts[0].Source, nil
}
