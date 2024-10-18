package store

import (
	"encoding/json"
	"errors"

	"github.com/valyentdev/ravel/pkg/core"
	"go.etcd.io/bbolt"
)

type InstanceEntry struct {
	Instance         core.Instance
	LastEvent        core.InstanceEvent
	UnreportedEvents []core.InstanceEvent
}

func getInstance(b *bbolt.Bucket) (core.Instance, error) {
	var instance core.Instance
	bytes := b.Get(instanceKey)
	if bytes == nil {
		return instance, errors.New("instance not found in bucket")
	}

	err := json.Unmarshal(bytes, &instance)
	if err != nil {
		return instance, err
	}

	return instance, nil
}

func putInstance(b *bbolt.Bucket, instance core.Instance) error {
	bytes, err := json.Marshal(instance)
	if err != nil {
		return err
	}

	return b.Put(instanceKey, bytes)
}

func putEvent(b *bbolt.Bucket, event core.InstanceEvent) error {
	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return b.Put(event.Id[:], bytes)
}

func (s *Store) LoadInstances() ([]InstanceEntry, error) {
	tx, err := s.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	instances := tx.Bucket(instancesBucket)
	if instances == nil {
		panic("instances bucket not found the Init function should have been called")
	}

	instanceList := []InstanceEntry{}
	err = instances.ForEachBucket(func(key []byte) error {
		bucket := instances.Bucket(key)
		var instance core.Instance

		instance, err := getInstance(bucket)
		if err != nil {
			return err
		}

		events := bucket.Bucket(instancesEventsBucket)
		if events == nil {
			return errors.New("events bucket not found")
		}

		cursor := events.Cursor()

		k, value := cursor.Last()
		if k == nil {
			return errors.New("no events found")
		}

		var lastEvent core.InstanceEvent

		if err = json.Unmarshal(value, &lastEvent); err != nil {
			return err
		}

		lastReported := bucket.Get(lastReportedEventIdKey)

		unreportedEvents, err := getUnreportedInstanceEvents(events, lastReported)
		if err != nil {
			return err
		}

		instanceList = append(instanceList, InstanceEntry{
			Instance:  instance,
			LastEvent: lastEvent,
			UnreportedEvents: unreportedEvents,
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

	instances := tx.Bucket(instancesBucket)

	ib, err := instances.CreateBucket([]byte(instance.Id))
	if err != nil {
		return err
	}

	err = putInstance(ib, instance)
	if err != nil {
		return err
	}

	events, err := ib.CreateBucket(instancesEventsBucket)
	if err != nil {
		return err
	}

	err = putEvent(events, event)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateInstance(instance core.Instance, event core.InstanceEvent) error {
	tx, err := s.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	instances := tx.Bucket(instancesBucket)

	ib := instances.Bucket([]byte(instance.Id))

	err = putInstance(ib, instance)
	if err != nil {
		return err
	}

	events := ib.Bucket(instancesEventsBucket)

	err = putEvent(events, event)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *Store) DestroyInstanceBucket(id string) error {
	tx, err := s.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	instances := tx.Bucket(instancesBucket)

	if err = instances.DeleteBucket([]byte(id)); err != nil {
		return err
	}

	return tx.Commit()
}
