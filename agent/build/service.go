package build

import (
	"context"
	"fmt"
	"sync"

	"github.com/alexisbouchez/ravel/api"
	"github.com/alexisbouchez/ravel/core/registry"
	"github.com/moby/buildkit/client"
)

// Config holds BuildKit service configuration
type Config struct {
	Enabled             bool   `json:"enabled" toml:"enabled"`
	Socket              string `json:"socket" toml:"socket"`
	MaxConcurrentBuilds int    `json:"max_concurrent_builds" toml:"max_concurrent_builds"`
}

// DefaultConfig returns the default BuildKit configuration
func DefaultConfig() Config {
	return Config{
		Enabled:             false,
		Socket:              "unix:///run/buildkit/buildkitd.sock",
		MaxConcurrentBuilds: 2,
	}
}

// Service manages container image builds using BuildKit
type Service struct {
	config     Config
	client     *client.Client
	store      *Store
	registries registry.RegistriesConfig

	builds     map[string]*buildState
	buildsLock sync.RWMutex

	sem chan struct{} // semaphore for limiting concurrent builds
}

// buildState tracks the state of an in-progress build
type buildState struct {
	build   *api.Build
	cancel  context.CancelFunc
	logFile string
}

// NewService creates a new BuildKit service
func NewService(config Config, store *Store, registries registry.RegistriesConfig) (*Service, error) {
	if !config.Enabled {
		return &Service{
			config: config,
			store:  store,
		}, nil
	}

	c, err := client.New(context.Background(), config.Socket)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to buildkit: %w", err)
	}

	maxBuilds := config.MaxConcurrentBuilds
	if maxBuilds <= 0 {
		maxBuilds = 2
	}

	return &Service{
		config:     config,
		client:     c,
		store:      store,
		registries: registries,
		builds:     make(map[string]*buildState),
		sem:        make(chan struct{}, maxBuilds),
	}, nil
}

// IsEnabled returns whether the build service is enabled
func (s *Service) IsEnabled() bool {
	return s.config.Enabled && s.client != nil
}

// Close closes the BuildKit client connection
func (s *Service) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

// GetBuild returns the current state of a build
func (s *Service) GetBuild(ctx context.Context, id string) (*api.Build, error) {
	// First check in-memory state for active builds
	s.buildsLock.RLock()
	state, ok := s.builds[id]
	s.buildsLock.RUnlock()

	if ok {
		return state.build, nil
	}

	// Fall back to persistent storage
	return s.store.GetBuild(ctx, id)
}

// ListBuilds returns all builds, optionally filtered by namespace
func (s *Service) ListBuilds(ctx context.Context, namespace string, limit int) ([]*api.Build, error) {
	return s.store.ListBuilds(ctx, namespace, limit)
}

// CancelBuild cancels an in-progress build
func (s *Service) CancelBuild(id string) error {
	s.buildsLock.RLock()
	state, ok := s.builds[id]
	s.buildsLock.RUnlock()

	if !ok {
		return fmt.Errorf("build not found or already completed")
	}

	if state.cancel != nil {
		state.cancel()
	}

	return nil
}
