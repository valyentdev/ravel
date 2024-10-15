package store

import (
	"github.com/valyentdev/ravel/internal/agent/structs"
)

func (s *Store) LoadReservations() ([]structs.Reservation, error) {
	tx, err := s.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	reservations, err := tx.Bucket(reservationsBucket)
	if err != nil {
		return nil, err
	}

	reservationList := []structs.Reservation{}

	cursor := reservations.Cursor()

	for k, _ := cursor.First(); k != nil; k, _ = cursor.Next() {
		reservation := structs.Reservation{}
		err = reservations.Get(k, &reservation)
		if err != nil {
			return nil, err
		}
		reservationList = append(reservationList, reservation)
	}
	return reservationList, err
}

func (store *Store) DeleteReservation(id string) error {
	tx, err := store.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	reservations, err := tx.Bucket(reservationsBucket)
	if err != nil {
		return err
	}

	if err = reservations.Delete([]byte(id)); err != nil {
		return err
	}

	return nil
}

func (store *Store) PutReservation(r structs.Reservation) error {
	tx, err := store.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	reservations, err := tx.Bucket(reservationsBucket)
	if err != nil {
		return err
	}

	if err = reservations.Put([]byte(r.Id), r); err != nil {
		return err
	}

	return tx.Commit()
}
