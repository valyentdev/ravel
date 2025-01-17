package runtime


import "github.com/valyentdev/ravel/runtime/disks"

func (r *Runtime) GetDisk(id string) (*disks.Disk, error) {
	return r.disks.GetDisk(id)
}

func (r *Runtime) ListDisks() ([]disks.Disk, error) {
	return r.disks.ListDisks()
}

func (r *Runtime) DestroyDisk(id string) error {
	return r.disks.DestroyDisk(id)
}

func (r *Runtime) CreateDisk(id string, sizeMB uint64) (*disks.Disk, error) {
	return r.disks.CreateDisk(id, sizeMB)
}
