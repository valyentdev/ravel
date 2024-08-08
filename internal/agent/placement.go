package agent

import (
	"context"
	"log/slog"

	"github.com/valyentdev/ravel/internal/placement"
	"github.com/valyentdev/ravel/pkg/core"
)

func (a *Agent) startPlacementHandler() error {
	l := placement.NewListener(a.nc)
	max := a.reservations.Max()

	err := l.HandleMachinePlacementRequest(context.Background(), a.config.Region, func(msg *placement.MachinePlacementRequest, answer placement.AnswerFunc) {
		slog.Info("Received placement request", "request", msg)
		_, before, after, err := a.reservations.CreateReservation(context.Background(), msg.ReservationId, core.Resources{
			Cpus:   msg.Cpus,
			Memory: msg.Memory,
		}, core.ReservationStatusDangling)

		if err != nil {
			slog.Error("Failed to create reservation", "error", err)
			return
		}

		slog.Info("Reservation created", "before", before, "after", after)
		answer(&placement.MachinePlacementResponse{
			NodeId:          a.nodeId,
			Allocatable:     max,
			AllocatedBefore: before,
			AllocatedAfter:  after,
		})

	})
	if err != nil {
		return err
	}

	return nil
}
