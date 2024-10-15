package reservations

import (
	"time"

	"github.com/valyentdev/ravel/internal/agent/structs"
)

func (rs *ReservationService) gc(id string) {
	time.Sleep(10 * time.Second)
	rs.lock.Lock()
	defer rs.lock.Unlock()

	reservation, ok := rs.reservations[id]
	if !ok {
		return
	}

	if reservation.Status != structs.ReservationStatusDangling {
		return
	}

	subnet := reservation.LocalIPV4Subnet.LocalConfig().Network

	if err := rs.localSubnetAllocator.Release(subnet); err != nil {
		return
	}

	delete(rs.reservations, id)

	rs.current = rs.current.Sub(reservation.Resources)
}
