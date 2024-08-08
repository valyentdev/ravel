package reservations

import (
	"context"
	"log/slog"
	"time"
)

func (r *ReservationService) StartGarbageCollection(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.gc(ctx)
		}
	}

}

func (r *ReservationService) gc(ctx context.Context) {
	r.lock.Lock()
	defer r.lock.Unlock()

	err := r.store.GCReservations(ctx, time.Now().Add(-10*time.Second))
	if err != nil {
		slog.Error("failed to garbage collect reservations", "err", err)
	}

}
