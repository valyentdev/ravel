package state

import (
	"github.com/alexisbouchez/ravel/core/cluster"
	"github.com/alexisbouchez/ravel/ravel/state/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type State struct {
	clusterState cluster.ClusterState
	db           *db.DB
	Queries      *db.Queries
}

func New(pgxpool *pgxpool.Pool, clusterState cluster.ClusterState) *State {
	database := db.New(pgxpool)
	return &State{
		clusterState: clusterState,
		db:           database,
		Queries:      database.Queries,
	}
}
