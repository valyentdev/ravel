package store

import (
	"encoding/json"

	"github.com/alexisbouchez/ravel/api"
	"go.etcd.io/bbolt"
)

func assertEventsBucketExists(bucket *bbolt.Bucket) error {
	if bucket == nil {
		return ErrBucketNotFound
	}
	return nil
}

func (s *Store) DeleteMachineInstanceEvent(eventId string) error {
	tx, err := s.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	events := tx.Bucket(eventsBucket)
	if err := assertEventsBucketExists(events); err != nil {
		return err
	}

	if err = events.Delete([]byte(eventId)); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) PutMachineInstanceEvent(event api.MachineEvent) error {
	tx, err := s.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	events := tx.Bucket(eventsBucket)
	if err := assertEventsBucketExists(events); err != nil {
		return err
	}

	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	if err = events.Put([]byte(event.Id), bytes); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) LoadMachineInstanceEvents() ([]api.MachineEvent, error) {
	tx, err := s.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	events := tx.Bucket(eventsBucket)
	if err := assertEventsBucketExists(events); err != nil {
		return nil, err
	}

	var machineEvents []api.MachineEvent
	err = events.ForEach(func(k, v []byte) error {
		var event api.MachineEvent
		if err := json.Unmarshal(v, &event); err != nil {
			return err
		}
		machineEvents = append(machineEvents, event)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return machineEvents, nil
}
