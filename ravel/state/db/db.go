package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valyentdev/ravel/internal/dbutil"
)

type Queries struct {
	db dbutil.PGXDBTX
}

type DB struct {
	pool *pgxpool.Pool
	*Queries
}

func New(pool *pgxpool.Pool) *DB {
	return &DB{
		pool: pool,
		Queries: &Queries{
			db: pool,
		},
	}
}

type Transaction struct {
	tx pgx.Tx
	*Queries
}

func (db *DB) BeginTx(ctx context.Context) (*Transaction, error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &Transaction{
		tx: tx,
		Queries: &Queries{
			db: tx,
		},
	}, nil
}

func (tx *Transaction) Commit(ctx context.Context) error {
	return tx.tx.Commit(ctx)
}

func (tx *Transaction) Rollback(ctx context.Context) error {
	return tx.tx.Rollback(ctx)
}
