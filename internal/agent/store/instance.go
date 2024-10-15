package store

import (
	"errors"

	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/helper/superbolt"
)

type InstanceEntry struct {
	Instance  core.Instance
	LastEvent core.InstanceEvent
}

func (s *Store) LoadInstances() ([]InstanceEntry, error) {
	tx, err := s.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	instances, err := tx.Bucket(instancesBucket)
	if err != nil {
		return nil, err
	}

	instanceList := []InstanceEntry{}
	err = instances.ForEachBucket(func(key []byte, bucket *superbolt.Bucket) error {
		var instance core.Instance
		var lastEvent core.InstanceEvent
		err := bucket.Get(instanceKey, &instance)
		if err != nil {
			return err
		}

		events, err := bucket.Bucket(instancesEventsBucket)
		if err != nil {
			return err
		}

		cursor := events.Cursor()

		k, _ := cursor.Last()
		if k == nil {
			return errors.New("no events found")
		}

		err = events.Get(k, &lastEvent)
		if err != nil {
			return err
		}

		instanceList = append(instanceList, InstanceEntry{
			Instance:  instance,
			LastEvent: lastEvent,
		})
		return nil
	})

	return instanceList, err
}

func (s *Store) CreateInstance(instance core.Instance, event core.InstanceEvent) error {
	tx, err := s.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	instances, err := tx.Bucket(instancesBucket)
	if err != nil {
		return err
	}

	ib, err := instances.CreateBucket([]byte(instance.Id))
	if err != nil {
		return err
	}

	err = ib.Put(instanceKey, instance)
	if err != nil {
		return err
	}

	events, err := ib.CreateBucket(instancesEventsBucket)
	if err != nil {
		return err
	}

	err = events.Put(event.Id[:], event)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (store *Store) UpdateInstance(instance core.Instance, event core.InstanceEvent) error {
	tx, err := store.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	instances, err := tx.Bucket(instancesBucket)
	if err != nil {
		return err
	}

	ib, err := instances.Bucket([]byte(instance.Id))
	if err != nil {
		return err
	}

	err = ib.Put(instanceKey, instance)
	if err != nil {
		return err
	}

	events, err := ib.Bucket(instancesEventsBucket)
	if err != nil {
		return err
	}

	err = events.Put(event.Id[:], event)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (store *Store) DestroyInstanceBucket(id string) error {
	tx, err := store.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	instances, err := tx.Bucket(instancesBucket)
	if err != nil {
		return err
	}

	if err = instances.DeleteBucket([]byte(id)); err != nil {
		return err
	}

	return tx.Commit()
}
