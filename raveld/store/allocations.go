package store

import (
	"encoding/json"

	"github.com/alexisbouchez/ravel/agent/structs"
	"go.etcd.io/bbolt"
)

func assertAllocationsBucketExists(bucket *bbolt.Bucket) error {
	if bucket == nil {
		return ErrBucketNotFound
	}
	return nil
}

func (s *Store) LoadAllocations() ([]structs.Allocation, error) {
	tx, err := s.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	allocations := tx.Bucket(allocationsBucket)
	if err := assertAllocationsBucketExists(allocations); err != nil {
		return nil, err
	}

	allocationsList := []structs.Allocation{}

	cursor := allocations.Cursor()
	for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
		var allocation structs.Allocation

		if err := json.Unmarshal(v, &allocation); err != nil {
			return nil, err
		}

		allocationsList = append(allocationsList, allocation)
	}
	return allocationsList, err
}

func (s *Store) DeleteAllocation(id string) error {
	tx, err := s.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	allocations := tx.Bucket(allocationsBucket)
	if err := assertAllocationsBucketExists(allocations); err != nil {
		return err
	}

	if err = allocations.Delete([]byte(id)); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func putAllocation(b *bbolt.Bucket, a structs.Allocation) error {
	bytes, err := json.Marshal(a)
	if err != nil {
		return err
	}

	return b.Put([]byte(a.Id), bytes)
}

func (s *Store) PutAllocation(a structs.Allocation) error {
	tx, err := s.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	allocations := tx.Bucket(allocationsBucket)
	if err := assertAllocationsBucketExists(allocations); err != nil {
		return err
	}

	if err = putAllocation(allocations, a); err != nil {
		return err
	}

	return tx.Commit()
}
