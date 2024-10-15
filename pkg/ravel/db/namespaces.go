package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/valyentdev/ravel/internal/dbutil"
	"github.com/valyentdev/ravel/pkg/core"
)

func scanNamespace(s dbutil.Scannable) (*core.Namespace, error) {
	var n core.Namespace
	var createdAt pgtype.Timestamp
	err := s.Scan(&n.Name, &createdAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, core.NewNotFound("namespace not found")
		}
		return nil, err
	}

	n.CreatedAt = createdAt.Time
	return &n, err
}

func (q *Queries) CreateNamespace(ctx context.Context, namespace core.Namespace) error {
	_, err := q.db.Exec(ctx, `INSERT INTO namespaces (name, created_at) VALUES ($1, $2)`, namespace.Name, namespace.CreatedAt)
	if err != nil {
		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) {
			if pgerr.ConstraintName == "namespaces_pkey" {
				return core.NewAlreadyExists("namespace already exists")
			}
		}
		return err
	}
	return nil
}

func (q *Queries) ListNamespaces(ctx context.Context) ([]core.Namespace, error) {
	rows, err := q.db.Query(ctx, `SELECT name, created_at FROM namespaces`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	namespaces := []core.Namespace{}
	for rows.Next() {
		namespace, err := scanNamespace(rows)
		if err != nil {
			return nil, err
		}
		namespaces = append(namespaces, *namespace)
	}

	return namespaces, nil
}

func (q *Queries) GetNamespace(ctx context.Context, name string) (*core.Namespace, error) {
	row := q.db.QueryRow(ctx, `SELECT name, created_at FROM namespaces WHERE name = $1`, name)
	namespace, err := scanNamespace(row)
	if err != nil {
		return nil, err
	}
	return namespace, nil
}

func (q *Queries) DestroyNamespace(ctx context.Context, name string) error {
	_, err := q.db.Exec(ctx, `DELETE FROM namespaces WHERE name = $1`, name)
	if err != nil {
		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) {
			if pgerr.ConstraintName == "fleets_namespace_fkey" {
				return core.NewFailedPrecondition("namespace still has fleets")
			}
		}
		return err
	}
	return nil
}
