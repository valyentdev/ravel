package state

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valyentdev/ravel/core/cluster"
	"github.com/valyentdev/ravel/ravel/state/db"
)

type State struct {
	clusterState cluster.ClusterState
	db           *db.DB
}

func New(pgxpool *pgxpool.Pool, clusterState cluster.ClusterState) *State {
	db := db.New(pgxpool)
	return &State{
		clusterState: clusterState,
		db:           db,
	}
}
