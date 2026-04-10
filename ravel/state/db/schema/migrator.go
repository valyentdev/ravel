package schema

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/tern/v2/migrate"
)

type Migration struct {
	Name string
	Up   string
	Down string
}

type Migrator struct {
	migrations []Migration
	tern       *migrate.Migrator
}

func NewMigrator(conn *pgx.Conn) (*Migrator, error) {
	migrator, err := migrate.NewMigrator(context.Background(), conn, "ravel_schema")
	if err != nil {
		return nil, err
	}

	migrations := getMigrations()

	for _, m := range migrations {
		migrator.AppendMigration(m.Name, m.Up, m.Down)
	}

	return &Migrator{tern: migrator, migrations: migrations}, nil
}

func (m *Migrator) Migrate(ctx context.Context) error {
	return m.tern.Migrate(ctx)
}

func (m *Migrator) MigrateTo(ctx context.Context, version int32) error {
	return m.tern.MigrateTo(ctx, version)
}

func (m *Migrator) Info(ctx context.Context) (current int32, total int32, migrations []Migration, err error) {
	current, err = m.tern.GetCurrentVersion(ctx)
	if err != nil {
		return
	}

	total = int32(len(m.migrations))
	migrations = m.migrations

	return
}
