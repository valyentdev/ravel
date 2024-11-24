package corrosion

import "github.com/valyentdev/corroclient"

type CorrosionClusterState struct {
	corroclient *corroclient.CorroClient // use the corrosion http api for subscriptions
}

func New(config corroclient.Config) *CorrosionClusterState {
	client := corroclient.NewCorroClient(config)
	return &CorrosionClusterState{
		corroclient: client,
	}
}
