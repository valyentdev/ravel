package agent

import (
	"context"
	"log/slog"

	"github.com/valyentdev/ravel/core/cluster/placement"
)

func (a *Agent) stopPlacementHandler() {

}

func (a *Agent) startPlacementHandler() error {
	l := placement.NewListener(a.nc)
	max := a.allocator.Max()

	err := l.HandleMachinePlacementRequest(
		context.Background(),
		a.config.Region,
		func(msg *placement.PlacementRequest) *placement.PlacementResponse {
			slog.Debug("Received placement request", "request", msg)
			_, before, after, err := a.allocator.CreateAllocation(msg.AllocationId, msg.Resources)
			if err != nil {
				slog.Error("Failed to create reservation", "error", err)
				return nil
			}

			slog.Debug("Allocation created", "before", before, "after", after)
			return &placement.PlacementResponse{
				NodeId:          a.node.Id(),
				Allocatable:     max,
				AllocatedBefore: before,
				AllocatedAfter:  after,
			}
		})
	if err != nil {
		return err
	}

	return nil
}
