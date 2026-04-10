package runtime

import (
	"context"

	"github.com/alexisbouchez/ravel/runtime/disks"
)

func (r *Runtime) GetDisk(id string) (*disks.Disk, error) {
	return r.disks.GetDisk(id)
}

func (r *Runtime) ListDisks() ([]disks.Disk, error) {
	return r.disks.ListDisks()
}

func (r *Runtime) DestroyDisk(id string) error {
	return r.disks.DestroyDisk(id)
}

func (r *Runtime) CreateDisk(ctx context.Context, id string, sizeMB uint64) (*disks.Disk, error) {
	return r.disks.CreateDisk(ctx, id, sizeMB)
}
