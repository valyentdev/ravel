package store

import "github.com/valyentdev/ravel/pkg/helper/superbolt"

type Store struct {
	db *superbolt.DB
}

/**
Schema:
instances/
	  <instance_id>/
	  		instance -> core.Instance
			events/
				  <event_id> -> core.InstanceEvent
reservations/
	  <reservation_id> -> structs.Reservation
**/

var (
	instancesBucket       = []byte("instances")
	instanceKey           = []byte("instance")
	instancesEventsBucket = []byte("events")

	reservationsBucket = []byte("reservations")
)

func NewStore(path string) (*Store, error) {
	db, err := superbolt.Open(path, 0600, nil)
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

	_, err = tx.CreateBucketIfNotExists(reservationsBucket)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) Close() error {
	return s.db.Close()
}
