package disks

import (
	"os/exec"

	"github.com/mistifyio/go-zfs/v3"
)

type ZFSPool struct {
	pool string
}

func NewZFSPool(pool string) *ZFSPool {
	return &ZFSPool{
		pool: pool,
	}
}

type DevicePool interface {
	CreateDevice(id string, size uint64) (string, error)
	DeleteDevice(id string) error
	Snapshot(id, snapshot string) error
	DeleteSnapshot(id, snapshot string) error
}

func (z *ZFSPool) volumeName(id string) string {
	return z.pool + "/" + id
}

func (z *ZFSPool) snaphotName(id, snapshot string) string {
	return z.volumeName(id) + "@" + snapshot
}

func (z *ZFSPool) devPath(id string) string {
	return "/dev/zvol/" + z.volumeName(id)
}

func (z *ZFSPool) CreateDevice(id string, size uint64) (string, error) {
	_, err := zfs.CreateVolume(z.volumeName(id), size*1024*1024, nil)
	if err != nil {
		return "", err
	}
	exec.Command("zvol_wait").Run()

	return z.devPath(id), nil
}

func (z *ZFSPool) DeleteDevice(id string) error {
	dataset := zfs.Dataset{
		Name: z.volumeName(id),
	}
	return dataset.Destroy(zfs.DestroyDefault)
}

func (z *ZFSPool) Snapshot(id, snapshot string) error {
	dataset := zfs.Dataset{
		Name: z.volumeName(id),
	}
	_, err := dataset.Snapshot(snapshot, false)
	return err
}

func (z *ZFSPool) DeleteSnapshot(id, snapshot string) error {
	dataset := zfs.Dataset{
		Name: z.snaphotName(id, snapshot),
	}
	return dataset.Destroy(zfs.DestroyDefault)
}
