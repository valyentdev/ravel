package ravel

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/valyentdev/ravel/api/errdefs"
	"github.com/valyentdev/ravel/core/cluster"
)

type regions struct {
	regions map[string]struct{}
	mutex   sync.RWMutex
	state   cluster.ClusterState
}

func (r *regions) validate(region string) error {
	r.mutex.RLock()
	_, ok := r.regions[region]
	r.mutex.RUnlock()
	if !ok {
		return errdefs.NewNotFound("region not found")
	}
	return nil
}

func (r *regions) startPolling(interval time.Duration) {
	for {
		regions, err := r.state.ListRegions(context.Background())
		if err != nil {
			slog.Error("failed to list regions", "err", err)
			continue
		}

		newRegions := make(map[string]struct{})
		for _, region := range regions {
			newRegions[region] = struct{}{}
		}

		r.mutex.Lock()
		r.regions = newRegions
		r.mutex.Unlock()

		time.Sleep(interval)

	}
}
