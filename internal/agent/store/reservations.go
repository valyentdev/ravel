package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/ravelerrors"
)

var reservationFields = []string{
	"id",
	"cpus",
	"memory",
	"local_ipv4_subnet",
	"status",
	"created_at",
}

var reservationColumns = allColumns(reservationFields)
var reservationPlaceholders = placeholders(len(reservationFields))

func scanReservation(row scannable) (core.Reservation, error) {
	var reservation core.Reservation

	err := row.Scan(
		&reservation.Id,
		&reservation.Cpus,
		&reservation.Memory,
		&reservation.LocalIPV4Subnet,
		&reservation.Status,
		&reservation.CreatedAt,
	)

	return reservation, err
}

func (s *Queries) ListReservations(ctx context.Context) ([]core.Reservation, error) {
	rows, err := s.db.QueryContext(ctx, fmt.Sprintf("SELECT %s FROM reservations", reservationColumns))
	if err != nil {
		if err == sql.ErrNoRows {
			return []core.Reservation{}, nil
		}
		return nil, err
	}

	defer rows.Close()

	reservations := []core.Reservation{}

	for rows.Next() {
		reservation, err := scanReservation(rows)
		if err != nil {
			return nil, err
		}

		reservations = append(reservations, reservation)
	}

	return reservations, nil
}

func (s *Queries) ListDanglingReservations(ctx context.Context) ([]core.Reservation, error) {
	rows, err := s.db.QueryContext(ctx, fmt.Sprintf("SELECT %s FROM reservations WHERE status = 'dangling'", reservationColumns))
	if err != nil {
		if err == sql.ErrNoRows {
			return []core.Reservation{}, nil
		}
		return nil, err
	}

	defer rows.Close()

	reservations := []core.Reservation{}

	for rows.Next() {
		reservation, err := scanReservation(rows)
		if err != nil {
			return nil, err
		}

		reservations = append(reservations, reservation)
	}

	return reservations, nil
}

func (s *Queries) GetReservation(ctx context.Context, id string) (res core.Reservation, err error) {
	row := s.db.QueryRowContext(ctx, fmt.Sprintf("SELECT %s FROM reservations WHERE id = ? LIMIT 1", reservationColumns), id)

	res, err = scanReservation(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, ravelerrors.NewNotFound("reservation not found")
		}

		return
	}
	return
}

type CreateReservationParams struct {
	Id        string
	MilliCpus int64
	Memory    int64
}

func (s *Queries) CreateReservation(ctx context.Context, params core.Reservation) error {
	_, err := s.db.ExecContext(
		ctx,
		fmt.Sprintf("INSERT INTO reservations (%s) VALUES (%s)", reservationColumns, reservationPlaceholders),
		params.Id,
		params.Cpus,
		params.Memory,
		params.LocalIPV4Subnet,
		params.Status,
		params.CreatedAt,
	)
	return err
}

func (s *Queries) UpdateReservation(ctx context.Context, params core.Reservation) error {
	_, err := s.db.ExecContext(
		ctx,
		"UPDATE reservations SET cpus = ?, memory = ?, status = ?, local_ipv4_subnet = ?, created_at = ? WHERE id = ?",
		params.Cpus,
		params.Memory,
		params.Status,
		params.LocalIPV4Subnet,
		params.CreatedAt,
		params.Id,
	)
	return err
}

func (s *Queries) ConfirmReservation(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(
		ctx,
		"UPDATE reservations SET status = 'confirmed' WHERE id = ?",
		id,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *Queries) DeleteReservation(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(
		ctx,
		"DELETE FROM reservations WHERE id = ?",
		id,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *Queries) GCReservations(ctx context.Context, cutoff time.Time) error {
	_, err := s.db.ExecContext(
		ctx,
		"DELETE FROM reservations WHERE created_at < ? AND status = 'dangling'",
		cutoff,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *Queries) GetReservedResources(ctx context.Context) (core.Resources, error) {
	row := s.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(cpus),0), COALESCE(SUM(memory),0) FROM reservations WHERE status = 'confirmed'")
	if row.Err() != nil {
		return core.Resources{}, row.Err()
	}

	var resources core.Resources
	err := row.Scan(&resources.Cpus, &resources.Memory)
	if err != nil {
		return core.Resources{}, err
	}
	return resources, nil
}
