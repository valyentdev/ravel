package corrosion

import (
	"context"

	"github.com/valyentdev/corroclient"
	"github.com/valyentdev/ravel/core/cluster"
)

type CorrosionClusterState struct {
	corroclient *corroclient.CorroClient // use the corrosion http api for subscriptions
}

var _ cluster.ClusterState = (*CorrosionClusterState)(nil)

func New(config corroclient.Config) *CorrosionClusterState {
	client := corroclient.NewCorroClient(config)
	return &CorrosionClusterState{
		corroclient: client,
	}
}

// DestroyNamespaceData implements cluster.ClusterState.
func (c *CorrosionClusterState) DestroyNamespaceData(ctx context.Context, namespace string) error {
	result, err := c.corroclient.Exec(ctx, []corroclient.Statement{
		{
			Query:  `DELETE FROM gateways WHERE namespace = $1`, // should be useless but just in case
			Params: []any{namespace},
		},
		{
			Query:  `DELETE FROM machine_versions WHERE namespace = $1`,
			Params: []any{namespace},
		},
		{
			Query:  `DELETE FROM instances WHERE namespace = $1`,
			Params: []any{namespace},
		},
		{
			Query:  `DELETE FROM machines WHERE namespace = $1`,
			Params: []any{namespace},
		},
	})
	if err != nil {
		return err
	}

	if err := result.Results[0].Err(); err != nil {
		return err
	}

	return nil
}
