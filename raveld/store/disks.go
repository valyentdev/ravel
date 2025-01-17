package store

import (
	"encoding/json"

	"github.com/valyentdev/ravel/api/errdefs"
	"github.com/valyentdev/ravel/runtime/disks"
	"go.etcd.io/bbolt"
)

func assertDisksBucketExists(bucket *bbolt.Bucket) {
	if bucket == nil {
		panic("disks bucket not found the Init function should have been called")
	}
}

type DiskTX struct {
	tx     *bbolt.Tx
	bucket *bbolt.Bucket
}

func (dtx *DiskTX) Commit() error {
	return dtx.tx.Commit()
}

func (dtx *DiskTX) Rollback() error {
	return dtx.tx.Rollback()
}

func (s *Store) BeginDiskTX(writable bool) (disks.DiskTX, error) {
	tx, err := s.db.Begin(writable)
	if err != nil {
		return nil, err
	}

	bucket := tx.Bucket(disksBucket)
	assertDisksBucketExists(bucket)

	return &DiskTX{tx: tx, bucket: bucket}, nil
}

func (s *DiskTX) ListDisks() ([]disks.Disk, error) {
	var result []disks.Disk
	err := s.bucket.ForEach(func(k, v []byte) error {
		var disk disks.Disk
		err := json.Unmarshal(v, &disk)
		if err != nil {
			return err
		}

		result = append(result, disk)
		return nil
	},
	)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *DiskTX) PutDisk(disk *disks.Disk) error {
	bucket := s.tx.Bucket(disksBucket)
	assertDisksBucketExists(bucket)

	bytes, err := json.Marshal(disk)
	if err != nil {
		return err
	}

	if err = bucket.Put([]byte(disk.Id), bytes); err != nil {
		return err
	}

	return nil
}

func (s *DiskTX) GetDisk(id string) (*disks.Disk, error) {
	bucket := s.tx.Bucket(disksBucket)
	assertDisksBucketExists(bucket)

	bytes := bucket.Get([]byte(id))
	if bytes == nil {
		return nil, errdefs.NewNotFound("disk not found")
	}

	var disk disks.Disk
	err := json.Unmarshal(bytes, &disk)
	if err != nil {
		return nil, err
	}

	return &disk, nil
}

func (s *DiskTX) DeleteDisk(id string) error {
	bucket := s.tx.Bucket(disksBucket)
	assertDisksBucketExists(bucket)
	if err := bucket.Delete([]byte(id)); err != nil {
		return err
	}

	return nil
}
