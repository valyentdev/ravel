package corrosion

import "github.com/valyentdev/corroclient"

type CorrosionClusterState struct {
	corroclient *corroclient.CorroClient // use the corrosion http api for subscriptions
}

func Connect(config corroclient.Config) (*CorrosionClusterState, error) {
	client := corroclient.NewCorroClient(config)

	return &CorrosionClusterState{
		corroclient: client,
	}, nil
}
