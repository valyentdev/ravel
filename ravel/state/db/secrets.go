package db

import (
	"context"
	"time"

	"github.com/alexisbouchez/ravel/api"
	"github.com/alexisbouchez/ravel/api/errdefs"
	"github.com/alexisbouchez/ravel/internal/dbutil"
	"github.com/jackc/pgx/v5"
)

func scanSecret(s dbutil.Scannable) (secret api.Secret, err error) {
	err = s.Scan(&secret.Id, &secret.Name, &secret.Namespace, &secret.CreatedAt, &secret.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			err = errdefs.NewNotFound("secret not found")
		}
		return
	}
	return
}

const createSecretQuery = `
INSERT INTO secrets (id, name, namespace, value, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
`

func (q *Queries) CreateSecret(ctx context.Context, id, name, namespace, value string) error {
	now := time.Now()
	_, err := q.db.Exec(ctx, createSecretQuery, id, name, namespace, value, now, now)
	if err != nil {
		return err
	}
	return nil
}

const getSecretQuery = `
SELECT id, name, namespace, created_at, updated_at
FROM secrets
WHERE namespace = $1 AND name = $2
`

func (q *Queries) GetSecret(ctx context.Context, namespace, name string) (api.Secret, error) {
	row := q.db.QueryRow(ctx, getSecretQuery, namespace, name)
	return scanSecret(row)
}

const getSecretValueQuery = `
SELECT value
FROM secrets
WHERE namespace = $1 AND name = $2
`

func (q *Queries) GetSecretValue(ctx context.Context, namespace, name string) (string, error) {
	var value string
	err := q.db.QueryRow(ctx, getSecretValueQuery, namespace, name).Scan(&value)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", errdefs.NewNotFound("secret not found")
		}
		return "", err
	}
	return value, nil
}

const listSecretsQuery = `
SELECT id, name, namespace, created_at, updated_at
FROM secrets
WHERE namespace = $1
ORDER BY created_at DESC
`

func (q *Queries) ListSecrets(ctx context.Context, namespace string) ([]api.Secret, error) {
	rows, err := q.db.Query(ctx, listSecretsQuery, namespace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var secrets []api.Secret
	for rows.Next() {
		secret, err := scanSecret(rows)
		if err != nil {
			return nil, err
		}
		secrets = append(secrets, secret)
	}

	return secrets, rows.Err()
}

const updateSecretQuery = `
UPDATE secrets
SET value = $1, updated_at = $2
WHERE namespace = $3 AND name = $4
`

func (q *Queries) UpdateSecret(ctx context.Context, namespace, name, value string) error {
	result, err := q.db.Exec(ctx, updateSecretQuery, value, time.Now(), namespace, name)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errdefs.NewNotFound("secret not found")
	}

	return nil
}

const deleteSecretQuery = `
DELETE FROM secrets
WHERE namespace = $1 AND name = $2
`

func (q *Queries) DeleteSecret(ctx context.Context, namespace, name string) error {
	result, err := q.db.Exec(ctx, deleteSecretQuery, namespace, name)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errdefs.NewNotFound("secret not found")
	}

	return nil
}
