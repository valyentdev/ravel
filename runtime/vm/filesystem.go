package vm

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/containerd/containerd/v2/client"
	"github.com/containerd/errdefs"
	"github.com/opencontainers/image-spec/identity"
)

func (b *Builder) removeRootFS(id string) error {
	ss := b.ctrd.SnapshotService(b.snapshotter)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := ss.Remove(ctx, rootFSName(id))
	if err != nil {
		if errdefs.IsNotFound(err) {
			return nil
		}

		return fmt.Errorf("failed to remove snapshot %q: %w", id, err)
	}

	return nil
}

func (b *Builder) prepareRootFS(ctx context.Context, id string, image client.Image) (rootfs string, err error) {
	rootfs, err = b.prepareContainerRootFS(ctx, id, image)
	if err != nil {
		return "", fmt.Errorf("failed to prepare rootfs for instance %q: %w", id, err)
	}
	defer func() {
		if err != nil {
			b.removeRootFS(id)
		}
	}()

	return rootfs, nil
}

func rootFSName(id string) string {
	return fmt.Sprintf("%s-%s", id, "rootfs")
}

func (b *Builder) prepareContainerRootFS(ctx context.Context, id string, image client.Image) (string, error) {
	slog.Debug("preparing rootfs for container", "id", id)
	diffIDs, err := image.RootFS(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get rootfs for image %q: %w", image.Name(), err)
	}

	parent := identity.ChainID(diffIDs).String()

	slog.Debug("preparing snapshot", "id", id, "parent", parent)

	ss := b.ctrd.SnapshotService(b.snapshotter)

	slog.Debug("preparing snapshot", "id", id, "parent", parent)
	mounts, err := ss.Prepare(context.Background(), rootFSName(id), parent)
	if err != nil {
		return "", fmt.Errorf("failed to prepare snapshot %q: %w", id, err)
	}

	if len(mounts) == 0 {
		return "", fmt.Errorf("no mounts found for container %q", id)
	}

	return mounts[0].Source, nil
}
