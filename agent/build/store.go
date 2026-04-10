package build

import (
	"context"
	"database/sql"
	"time"

	"github.com/alexisbouchez/ravel/api"
	_ "modernc.org/sqlite"
)

// Store handles persistence of build state
type Store struct {
	db *sql.DB
}

// NewStore creates a new build store from a database path
func NewStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	store := &Store{db: db}

	// Run migrations
	if err := store.Migrate(context.Background()); err != nil {
		db.Close()
		return nil, err
	}

	return store, nil
}

// Migrate creates the builds table if it doesn't exist
func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS builds (
			id TEXT PRIMARY KEY,
			namespace TEXT NOT NULL,
			image_name TEXT NOT NULL,
			tag TEXT NOT NULL,
			registry TEXT NOT NULL,
			full_image TEXT NOT NULL,
			status TEXT NOT NULL,
			digest TEXT,
			error TEXT,
			duration_ms INTEGER,
			created_at TEXT NOT NULL,
			completed_at TEXT
		)
	`)
	return err
}

// CreateBuild inserts a new build record
func (s *Store) CreateBuild(ctx context.Context, build *api.Build) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO builds (id, namespace, image_name, tag, registry, full_image, status, digest, error, duration_ms, created_at, completed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		build.Id,
		build.Namespace,
		build.ImageName,
		build.Tag,
		build.Registry,
		build.FullImage,
		build.Status,
		build.Digest,
		build.Error,
		build.DurationMs,
		build.CreatedAt.Format(time.RFC3339),
		nil,
	)
	return err
}

// UpdateBuild updates an existing build record
func (s *Store) UpdateBuild(ctx context.Context, build *api.Build) error {
	var completedAt *string
	if build.CompletedAt != nil {
		t := build.CompletedAt.Format(time.RFC3339)
		completedAt = &t
	}

	_, err := s.db.ExecContext(ctx, `
		UPDATE builds
		SET status = ?, digest = ?, error = ?, duration_ms = ?, completed_at = ?
		WHERE id = ?
	`,
		build.Status,
		build.Digest,
		build.Error,
		build.DurationMs,
		completedAt,
		build.Id,
	)
	return err
}

// GetBuild retrieves a build by ID
func (s *Store) GetBuild(ctx context.Context, id string) (*api.Build, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, namespace, image_name, tag, registry, full_image, status, digest, error, duration_ms, created_at, completed_at
		FROM builds
		WHERE id = ?
	`, id)

	return scanBuild(row)
}

// ListBuilds retrieves builds, optionally filtered by namespace
func (s *Store) ListBuilds(ctx context.Context, namespace string, limit int) ([]*api.Build, error) {
	var rows *sql.Rows
	var err error

	if namespace != "" {
		rows, err = s.db.QueryContext(ctx, `
			SELECT id, namespace, image_name, tag, registry, full_image, status, digest, error, duration_ms, created_at, completed_at
			FROM builds
			WHERE namespace = ?
			ORDER BY created_at DESC
			LIMIT ?
		`, namespace, limit)
	} else {
		rows, err = s.db.QueryContext(ctx, `
			SELECT id, namespace, image_name, tag, registry, full_image, status, digest, error, duration_ms, created_at, completed_at
			FROM builds
			ORDER BY created_at DESC
			LIMIT ?
		`, limit)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var builds []*api.Build
	for rows.Next() {
		build, err := scanBuildRows(rows)
		if err != nil {
			return nil, err
		}
		builds = append(builds, build)
	}

	return builds, rows.Err()
}

// DeleteBuild removes a build record
func (s *Store) DeleteBuild(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM builds WHERE id = ?`, id)
	return err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanBuild(row *sql.Row) (*api.Build, error) {
	var build api.Build
	var createdAt string
	var completedAt sql.NullString
	var digest sql.NullString
	var buildError sql.NullString
	var durationMs sql.NullInt64

	err := row.Scan(
		&build.Id,
		&build.Namespace,
		&build.ImageName,
		&build.Tag,
		&build.Registry,
		&build.FullImage,
		&build.Status,
		&digest,
		&buildError,
		&durationMs,
		&createdAt,
		&completedAt,
	)
	if err != nil {
		return nil, err
	}

	build.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	if completedAt.Valid {
		t, _ := time.Parse(time.RFC3339, completedAt.String)
		build.CompletedAt = &t
	}
	if digest.Valid {
		build.Digest = digest.String
	}
	if buildError.Valid {
		build.Error = buildError.String
	}
	if durationMs.Valid {
		build.DurationMs = durationMs.Int64
	}

	return &build, nil
}

func scanBuildRows(rows *sql.Rows) (*api.Build, error) {
	var build api.Build
	var createdAt string
	var completedAt sql.NullString
	var digest sql.NullString
	var buildError sql.NullString
	var durationMs sql.NullInt64

	err := rows.Scan(
		&build.Id,
		&build.Namespace,
		&build.ImageName,
		&build.Tag,
		&build.Registry,
		&build.FullImage,
		&build.Status,
		&digest,
		&buildError,
		&durationMs,
		&createdAt,
		&completedAt,
	)
	if err != nil {
		return nil, err
	}

	build.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	if completedAt.Valid {
		t, _ := time.Parse(time.RFC3339, completedAt.String)
		build.CompletedAt = &t
	}
	if digest.Valid {
		build.Digest = digest.String
	}
	if buildError.Valid {
		build.Error = buildError.String
	}
	if durationMs.Valid {
		build.DurationMs = durationMs.Int64
	}

	return &build, nil
}
