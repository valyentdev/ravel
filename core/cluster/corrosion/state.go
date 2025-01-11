package corrosion

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valyentdev/corroclient"
	"github.com/valyentdev/ravel/core/cluster"
	"github.com/valyentdev/ravel/internal/dbutil"
)

type Queries struct {
	dbtx dbutil.PGXDBTX
}

var _ cluster.Queries = (*Queries)(nil)

type CorrosionClusterState struct {
	db          *pgxpool.Pool
	corroclient *corroclient.CorroClient // use the corrosion http api for subscriptions
	cluster.Queries
}

var _ cluster.ClusterState = (*CorrosionClusterState)(nil)

type TX struct {
	tx pgx.Tx
	*Queries
}

func (t *TX) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

func (t *TX) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

// BeginTx implements cluster.ClusterState.
func (c *CorrosionClusterState) BeginTx(ctx context.Context) (cluster.TX, error) {
	tx, err := c.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &TX{
		tx:      tx,
		Queries: &Queries{dbtx: tx},
	}, nil
}

func New(config corroclient.Config, pgurl string) (*CorrosionClusterState, error) {
	client := corroclient.NewCorroClient(config)

	pool, err := pgxpool.New(context.Background(), pgurl)
	if err != nil {
		return nil, err
	}

	return &CorrosionClusterState{
		corroclient: client,
		db:          pool,
		Queries:     &Queries{dbtx: pool},
	}, nil
}

// DestroyNamespaceData implements cluster.ClusterState.
func (c *Queries) DestroyNamespaceData(ctx context.Context, namespace string) error {
	result := c.dbtx.SendBatch(ctx, &pgx.Batch{
		QueuedQueries: []*pgx.QueuedQuery{
			{
				SQL:       `DELETE FROM gateways WHERE namespace = $1`, // should be useless but just in case
				Arguments: []any{namespace},
			},
			{
				SQL:       `DELETE FROM machine_versions WHERE namespace = $1`,
				Arguments: []any{namespace},
			},
			{
				SQL:       `DELETE FROM instances WHERE namespace = $1`,
				Arguments: []any{namespace},
			},
			{
				SQL:       `DELETE FROM machines WHERE namespace = $1`,
				Arguments: []any{namespace},
			},
		},
	})

	err := result.Close()
	if err != nil {
		return err
	}

	return nil
}
