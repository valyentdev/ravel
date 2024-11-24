package ravel

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/valyentdev/ravel/core/cluster"
	"github.com/valyentdev/ravel/core/cluster/corrosion"
	"github.com/valyentdev/ravel/core/config"
	"github.com/valyentdev/ravel/ravel/db"
	"github.com/valyentdev/ravel/ravel/orchestrator"
	"github.com/valyentdev/ravel/ravel/state"
)

type Ravel struct {
	nc             *nats.Conn
	s              *orchestrator.Orchestrator
	db             *db.DB
	clusterState   cluster.ClusterState
	state          *state.State
	vcpusTemplates map[string]config.MachineResourcesTemplates
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

	clusterstate := corrosion.New(config.Corrosion.Config())
	if err != nil {
		return nil, err
	}

	s := orchestrator.New(nc, clusterstate)

	db := db.New(pgpool)

	return &Ravel{
		nc:             nc,
		s:              s,
		db:             db,
		clusterState:   clusterstate,
		state:          state.New(clusterstate, db),
		vcpusTemplates: config.Server.MachineTemplates,
	}, nil
}
