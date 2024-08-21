package filesystems

import (
	"context"
	"fmt"

	"github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/pkg/oci"
	"github.com/containerd/errdefs"
)

type ContainerFSBuilder struct {
	ctrd *client.Client
}

func NewContainerFSBuilder(containerdClient *client.Client) (*ContainerFSBuilder, error) {
	b := &ContainerFSBuilder{
		ctrd: containerdClient,
	}

	return b, nil
}

func (b *ContainerFSBuilder) CleanupFilesystems(id string) error {
	err := b.deleteContainerRootFS(id)
	if err != nil {
		return fmt.Errorf("failed to delete rootfs for machine %q: %w", id, err)
	}
	return nil
}

func (b *ContainerFSBuilder) PrepareFilesystems(id string, image client.Image) (rootfs string, err error) {
	rootfs, err = b.prepareContainerRootFS(id, image)
	if err != nil {
		return "", fmt.Errorf("failed to prepare rootfs for machine %q: %w", id, err)
	}
	defer func() {
		if err != nil {
			b.CleanupFilesystems(id)
		}
	}()

	return rootfs, nil
}

func (b *ContainerFSBuilder) createContainer(id string, image client.Image) error {
	_, err := b.ctrd.NewContainer(
		context.Background(),
		id,
		client.WithImage(image),
		client.WithSnapshotter("devmapper"),
		client.WithSpec(&oci.Spec{}),
		client.WithNewSnapshot(id, image),
	)

	return err
}

func (b *ContainerFSBuilder) prepareContainerRootFS(id string, image client.Image) (string, error) {
	err := b.createContainer(id, image)
	if err != nil {
		if !errdefs.IsAlreadyExists(err) {
			return "", fmt.Errorf("failed to create container %q: %w", id, err)
		}
		err = b.deleteContainerRootFS(id)
		if err != nil {
			return "", fmt.Errorf("failed to delete existing container %q: %w", id, err)
		}

		err = b.createContainer(id, image)
		if err != nil {
			return "", fmt.Errorf("failed to create container %q: %w", id, err)
		}
	}

	// Get the mounts for the container snapshot
	mounts, err := b.ctrd.SnapshotService("devmapper").Mounts(context.Background(), id)
	if err != nil {
		return "", fmt.Errorf("failed to get mounts %q: %w", id, err)
	}

	if len(mounts) == 0 {
		return "", fmt.Errorf("no mounts found for container %q", id)
	}

	return mounts[0].Source, nil
}

func (r *ContainerFSBuilder) deleteContainerRootFS(id string) error {
	err := r.ctrd.ContainerService().Delete(context.Background(), id)
	if err != nil && !errdefs.IsNotFound(err) {
		return fmt.Errorf("failed to delete container %q: %w", id, err)
	}

	snapshotter := r.ctrd.SnapshotService("devmapper")

	err = snapshotter.Remove(context.Background(), id)
	if err != nil && !errdefs.IsNotFound(err) {
		return fmt.Errorf("failed to remove snapshot %q: %w", id, err)
	}

	return nil
}
