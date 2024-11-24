package store

import (
	"encoding/json"
	"log/slog"

	"github.com/valyentdev/ravel/core/instance"
	"go.etcd.io/bbolt"
)

func assertInstancesBucketExists(bucket *bbolt.Bucket) {
	if bucket == nil {
		panic("instances bucket not found the Init function should have been called")
	}
}

var _ instance.InstanceStore = (*Store)(nil)

func (s *Store) LoadInstances() ([]instance.Instance, error) {
	tx, err := s.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	instances := tx.Bucket(instancesBucket)
	assertInstancesBucketExists(instances)

	instanceList := []instance.Instance{}
	err = instances.ForEach(func(k, v []byte) error {
		var instance instance.Instance
		if err := json.Unmarshal(v, &instance); err != nil {
			slog.Error("failed to unmarshal instance", "err", err)
			return nil
		}
		instanceList = append(instanceList, instance)
		return nil
	})

	return instanceList, err
}

func (s *Store) PutInstance(instance instance.Instance) error {
	tx, err := s.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	instances := tx.Bucket(instancesBucket)
	assertInstancesBucketExists(instances)

	bytes, err := json.Marshal(instance)
	if err != nil {
		return err
	}

	if err = instances.Put([]byte(instance.Id), bytes); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) DeleteInstance(id string) error {
	tx, err := s.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	instances := tx.Bucket(instancesBucket)
	assertInstancesBucketExists(instances)

	if err = instances.Delete([]byte(id)); err != nil {
		return err
	}

	return tx.Commit()
}
