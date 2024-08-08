package store

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/sqlite"
	_ "github.com/mattn/go-sqlite3"
	"github.com/valyentdev/ravel/db/agent/migrations"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

type Queries struct {
	db DBTX
}

type Store struct {
	*Queries
	db *sql.DB
}

func (s *Store) Close() error {
	return s.db.Close()
}

func NewStore() (*Store, error) {
	dbURL, err := url.Parse("sqlite:raveld.db")
	if err != nil {
		return nil, fmt.Errorf("failed to parse database url: %w", err)
	}

	migrator := dbmate.New(dbURL)
	migrator.FS = migrations.FS
	migrator.MigrationsDir = []string{"."}
	migrator.AutoDumpSchema = false
	if err = migrator.CreateAndMigrate(); err != nil {
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	db, err := sql.Open("sqlite3", "file:raveld.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return &Store{
		Queries: &Queries{db: db},
		db:      db,
	}, nil
}

type Transaction struct {
	*Queries
	tx *sql.Tx
}

func (s *Store) BeginTx(ctx context.Context) (*Transaction, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &Transaction{
		Queries: &Queries{db: tx},
		tx:      tx,
	}, nil
}

func (tx *Transaction) Commit() error {
	return tx.tx.Commit()
}

func (tx *Transaction) Rollback() error {
	return tx.tx.Rollback()
}
