package allocator

import (
	"time"

	"github.com/valyentdev/ravel/agent/structs"
)

func (rs *Allocator) gc(id string) {
	time.Sleep(10 * time.Second)
	rs.lock.Lock()
	defer rs.lock.Unlock()

	reservation, ok := rs.reservations[id]
	if !ok {
		return
	}

	if reservation.Status != structs.AllocationStatusDangling {
		return
	}

	delete(rs.reservations, id)

	rs.current = rs.current.Sub(reservation.Resources)
}
