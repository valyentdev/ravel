package store

import (
	"go.etcd.io/bbolt"
)

type Store struct {
	db *bbolt.DB
}

/**
Schema:
instances/
	  <instance_id> -> core.Instance
allocations/
	  <allocation_id> -> structs.Reservation
events/
	  <event_id> -> core.InstanceEvent
**/

var (
	machineInstancesBucket = []byte("machine_instances")
	instancesBucket        = []byte("instances")
	allocationsBucket      = []byte("allocations")
	eventsBucket           = []byte("events")
)

func NewStore(path string) (*Store, error) {
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Init() error {
	tx, err := s.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.CreateBucketIfNotExists(instancesBucket)
	if err != nil {
		return err
	}

	_, err = tx.CreateBucketIfNotExists(machineInstancesBucket)
	if err != nil {
		return err
	}

	_, err = tx.CreateBucketIfNotExists(allocationsBucket)
	if err != nil {
		return err
	}

	_, err = tx.CreateBucketIfNotExists(eventsBucket)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) Close() error {
	return s.db.Close()
}
