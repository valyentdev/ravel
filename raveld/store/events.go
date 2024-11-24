package store

import (
	"encoding/json"

	"github.com/oklog/ulid"
	"github.com/valyentdev/ravel/api"
	"go.etcd.io/bbolt"
)

func assertEventsBucketExists(bucket *bbolt.Bucket) {
	if bucket == nil {
		panic("events bucket not found the Init function should have been called")
	}
}

func (s *Store) DeleteMachineInstanceEvent(eventId ulid.ULID) error {
	tx, err := s.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	events := tx.Bucket(eventsBucket)
	assertEventsBucketExists(events)

	if err = events.Delete(eventId[:]); err != nil {
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
	assertEventsBucketExists(events)

	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	if err = events.Put(event.Id[:], bytes); err != nil {
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
	assertEventsBucketExists(events)

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
