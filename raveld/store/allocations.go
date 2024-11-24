package store

import (
	"encoding/json"

	"github.com/valyentdev/ravel/agent/structs"
	"go.etcd.io/bbolt"
)

func assertAllocationsBucketExists(bucket *bbolt.Bucket) {
	if bucket == nil {
		panic("allocations bucket not found the Init function should have been called")
	}
}

func (s *Store) LoadAllocations() ([]structs.Allocation, error) {
	tx, err := s.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	allocations := tx.Bucket(allocationsBucket)
	assertAllocationsBucketExists(allocations)

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

	if err = putAllocation(allocations, a); err != nil {
		return err
	}

	return tx.Commit()
}
