package disks

import (
	"context"
	"log/slog"
	"time"

	"github.com/valyentdev/ravel/api/errdefs"
	"github.com/valyentdev/ravel/core/validation"
)

type Disk struct {
	Id               string    `json:"id"`
	SizeMB           uint64    `json:"size_mb"`
	CreatedAt        time.Time `json:"created_at"`
	AttachedInstance string    `json:"attached_instance"`
	Path             string    `json:"path"`
}

type DiskSnapshot struct {
	Id        string    `json:"id"`
	DiskId    string    `json:"disk_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Store interface {
	BeginDiskTX(writable bool) (DiskTX, error)
}

type DiskTX interface {
	Commit() error
	Rollback() error
	ListDisks() ([]Disk, error)
	GetDisk(id string) (*Disk, error)
	PutDisk(disk *Disk) error
	DeleteDisk(id string) error
}

type Service struct {
	store Store
	pool  DevicePool
}

func NewService(store Store, pool DevicePool) *Service {
	return &Service{
		store: store,
		pool:  pool,
	}
}

func (s *Service) CreateDisk(id string, sizeMB uint64) (*Disk, error) {
	tx, err := s.store.BeginDiskTX(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	err = validation.ValidateObjectId(id)
	if err != nil {
		return nil, err
	}

	_, err = tx.GetDisk(id)
	if err == nil {
		return nil, errdefs.NewAlreadyExists("disk already exists")
	}

	path, err := s.pool.CreateDevice(id, sizeMB)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			time.Sleep(100 * time.Millisecond) // wait for the device to be released
			err = s.pool.DeleteDevice(id)
			if err != nil {
				slog.Error("failed to delete disk", "disk", id, "error", err)
			}
		}
	}()

	d := &Disk{
		Id:        id,
		SizeMB:    sizeMB,
		CreatedAt: time.Now(),
		Path:      path,
	}

	err = tx.PutDisk(d)
	if err != nil {
		return nil, err
	}

	slog.Debug("creating filesystem", "disk", id)

	err = MkfsEXT4(context.TODO(), path)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		slog.Error("failed to commit disk transaction", "error", err)
		return nil, err
	}

	return d, nil
}

func (s *Service) ListDisks() ([]Disk, error) {
	tx, err := s.store.BeginDiskTX(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	return tx.ListDisks()
}

func (s *Service) GetDisk(id string) (*Disk, error) {
	tx, err := s.store.BeginDiskTX(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	return tx.GetDisk(id)
}

func (s *Service) GetDisks(ids ...string) ([]Disk, error) {
	tx, err := s.store.BeginDiskTX(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	disks := make([]Disk, 0, len(ids))
	for _, id := range ids {
		d, err := tx.GetDisk(id)
		if err != nil {
			return nil, err
		}
		disks = append(disks, *d)
	}

	return disks, nil
}

func attachInstance(tx DiskTX, disk string, instance string) error {
	d, err := tx.GetDisk(disk)
	if err != nil {
		return err
	}

	if d.AttachedInstance != "" {
		return errdefs.NewAlreadyExists("disk is already attached")
	}

	d.AttachedInstance = instance

	err = tx.PutDisk(d)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) AttachInstance(instance string, disks ...string) error {
	if len(disks) == 0 {
		return nil
	}
	tx, err := s.store.BeginDiskTX(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, disk := range disks {
		err = attachInstance(tx, disk, instance)
		if err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func detachInstance(tx DiskTX, disk string) error {
	d, err := tx.GetDisk(disk)
	if err != nil {
		return err
	}

	if d.AttachedInstance == "" {
		return errdefs.NewNotFound("disk is not attached")
	}

	d.AttachedInstance = ""

	err = tx.PutDisk(d)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DetachInstance(disk ...string) error {
	if len(disk) == 0 {
		return nil
	}
	tx, err := s.store.BeginDiskTX(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, d := range disk {
		err = detachInstance(tx, d)
		if err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *Service) DestroyDisk(id string) error {
	tx, err := s.store.BeginDiskTX(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	d, err := tx.GetDisk(id)
	if err != nil {
		return err
	}

	if d.AttachedInstance != "" {
		return errdefs.NewAlreadyExists("disk is attached")
	}

	err = tx.DeleteDisk(id)
	if err != nil {
		return err
	}

	err = s.pool.DeleteDevice(id)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
