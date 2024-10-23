package store

import (
	"encoding/json"

	"github.com/valyentdev/ravel/internal/agent/structs"
	"go.etcd.io/bbolt"
)

func (s *Store) LoadReservations() ([]structs.Reservation, error) {
	tx, err := s.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	reservations := tx.Bucket(reservationsBucket)
	if reservations == nil {
		panic("reservations bucket not found the Init function should have been called")
	}

	reservationList := []structs.Reservation{}

	cursor := reservations.Cursor()
	for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
		var reservation structs.Reservation

		if err := json.Unmarshal(v, &reservation); err != nil {
			return nil, err
		}

		reservationList = append(reservationList, reservation)
	}
	return reservationList, err
}

func (s *Store) DeleteReservation(id string) error {
	tx, err := s.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	reservations := tx.Bucket(reservationsBucket)

	if err = reservations.Delete([]byte(id)); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func putReservation(b *bbolt.Bucket, r structs.Reservation) error {
	bytes, err := json.Marshal(r)
	if err != nil {
		return err
	}

	return b.Put([]byte(r.Id), bytes)
}

func (s *Store) PutReservation(r structs.Reservation) error {
	tx, err := s.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	reservations := tx.Bucket(reservationsBucket)

	if err = putReservation(reservations, r); err != nil {
		return err
	}

	return tx.Commit()
}
