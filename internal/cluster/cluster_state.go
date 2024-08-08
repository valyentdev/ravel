package cluster

import (
	"github.com/valyentdev/ravel/pkg/helper/corroclient"
)

type ClusterState struct {
	corroclient *corroclient.CorroClient // use the corrosion http api for subscriptions
}

func Connect(config corroclient.Config) (*ClusterState, error) {
	client := corroclient.NewCorroClient(config)

	return &ClusterState{
		corroclient: client,
	}, nil
}
