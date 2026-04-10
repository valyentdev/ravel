package ravel

import (
	"context"
	"crypto/tls"

	"github.com/alexisbouchez/ravel/core/cluster"
	"github.com/alexisbouchez/ravel/core/config"
	"github.com/alexisbouchez/ravel/ravel/orchestrator"
	"github.com/alexisbouchez/ravel/ravel/state"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
)

type Ravel struct {
	nc             *nats.Conn
	o              *orchestrator.Orchestrator
	State          *state.State
	pgpool         *pgxpool.Pool
	config         *config.RavelConfig
	vcpusTemplates map[string]config.MachineResourcesTemplates
}

func getClientTLSConfig(config config.RavelConfig) (*tls.Config, error) {
	if config.Server.TLS == nil {
		return nil, nil
	}

	cert, err := config.Server.TLS.LoadCert()
	if err != nil {
		return nil, err
	}

	ca, err := config.Server.TLS.LoadCA()
	if err != nil {
		return nil, err
	}

	insecureSkipVerify := config.Server.TLS.SkipVerifyServer

	return &tls.Config{
		RootCAs:            ca,
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: insecureSkipVerify,
	}, nil
}

func New(config config.RavelConfig) (*Ravel, error) {
	ctx := context.Background()

	pgpool, err := pgxpool.New(ctx, config.Server.PostgresURL)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			pgpool.Close()
		}
	}()

	natsOptions := []nats.Option{}
	if config.Nats.CredFile != "" {
		natsOptions = append(natsOptions, nats.UserCredentials(config.Nats.CredFile, config.Nats.CredFile), nats.MaxReconnects(-1))
	}

	nc, err := nats.Connect(config.Nats.Url, natsOptions...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			nc.Close()
		}
	}()

	clusterstate := cluster.NewNoop()

	tlsConfig, err := getClientTLSConfig(config)
	if err != nil {
		return nil, err
	}

	o := orchestrator.New(nc, clusterstate, tlsConfig)

	return &Ravel{
		nc:             nc,
		o:              o,
		State:          state.New(pgpool, clusterstate),
		vcpusTemplates: config.Server.MachineTemplates,
		pgpool:         pgpool,
		config:         &config,
	}, nil
}

func (r *Ravel) Start() error {
	return r.listenMachineEvents()
}

func (r *Ravel) Stop() error {
	r.nc.Close()
	r.pgpool.Close()
	return nil
}

// Ping checks connectivity to core dependencies.
// Returns an error if any critical dependency is unreachable.
func (r *Ravel) Ping(ctx context.Context) error {
	return r.pgpool.Ping(ctx)
}
