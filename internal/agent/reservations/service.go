package reservations

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/valyentdev/ravel/internal/agent/store"
	"github.com/valyentdev/ravel/internal/agent/structs"
	"github.com/valyentdev/ravel/internal/networking"
	"github.com/valyentdev/ravel/pkg/core"
)

var ErrNotEnoughResources = core.NewResourcesExhausted("not enough resources")

type ReservationService struct {
	store *store.Store
	max   core.Resources

	reservations map[string]structs.Reservation
	lock         sync.RWMutex
	current      core.Resources

	localSubnetAllocator *networking.BasicSubnetAllocator
}

func (rs *ReservationService) Max() core.Resources {
	return rs.max
}

func NewReservationService(store *store.Store, totalResources core.Resources) *ReservationService {
	subnetAllocator, err := networking.NewBasicSubnetAllocator(networking.IPNetPool{
		Pool: net.IPNet{
			IP:   net.ParseIP("172.18.0.0").To4(),
			Mask: net.CIDRMask(16, 32),
		},
		SubnetMask: net.CIDRMask(29, 32),
	})
	if err != nil {
		panic(err)
	}

	return &ReservationService{
		store:                store,
		max:                  totalResources,
		localSubnetAllocator: subnetAllocator,
		reservations:         make(map[string]structs.Reservation),
	}
}

func (rs *ReservationService) Init() error {
	rs.lock.Lock()
	defer rs.lock.Unlock()
	reservations, err := rs.store.LoadReservations()
	if err != nil {
		return err
	}

	for _, res := range reservations {
		subnet := res.LocalIPV4Subnet.LocalConfig().Network

		if err := rs.localSubnetAllocator.Allocate(subnet); err != nil {
			return err
		}

		rs.reservations[res.Id] = res

		if res.Status == structs.ReservationStatusConfirmed {
			rs.current = rs.current.Add(res.Resources)
		}
	}

	return nil
}

func (rs *ReservationService) CreateReservation(ctx context.Context, id string, res core.Resources) (reservation structs.Reservation, before core.Resources, after core.Resources, err error) {
	rs.lock.Lock()
	defer rs.lock.Unlock()

	localIPV4Subnet, err := rs.localSubnetAllocator.AllocateNext()
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			rs.localSubnetAllocator.Release(&localIPV4Subnet)
		}
	}()

	alloc := structs.Reservation{
		Id:              id,
		Resources:       res,
		LocalIPV4Subnet: networking.LocalIPV4Subnet(localIPV4Subnet.String()),
		Status:          structs.ReservationStatusDangling,
		CreatedAt:       time.Now(),
	}

	before = rs.current

	after = before.Add(res)

	if after.GT(rs.max) {
		err = ErrNotEnoughResources
		return
	}

	rs.current = after

	rs.reservations[id] = alloc

	go rs.gc(id)

	return
}

func (rs *ReservationService) GetReservation(id string) (structs.Reservation, error) {
	rs.lock.RLock()
	defer rs.lock.RUnlock()

	reservation, ok := rs.reservations[id]
	if !ok {
		return structs.Reservation{}, core.NewNotFound("reservation not found")
	}

	return reservation, nil
}

func (rs *ReservationService) DeleteReservation(id string) error {
	rs.lock.Lock()
	defer rs.lock.Unlock()

	reservation, ok := rs.reservations[id]
	if !ok {
		return nil
	}

	rs.current = rs.current.Sub(reservation.Resources)

	subnet := reservation.LocalIPV4Subnet.LocalConfig().Network

	if err := rs.localSubnetAllocator.Release(subnet); err != nil {
		return err
	}

	if err := rs.store.DeleteReservation(id); err != nil {
		return err
	}

	delete(rs.reservations, id)

	return nil
}

func (rs *ReservationService) ListReservations(ctx context.Context) []structs.Reservation {
	rs.lock.RLock()

	reservations := make([]structs.Reservation, len(rs.reservations))
	for _, res := range rs.reservations {
		reservations = append(reservations, res)
	}
	rs.lock.RUnlock()
	return reservations
}

func (rs *ReservationService) ConfirmReservation(ctx context.Context, id string) (structs.Reservation, error) {
	rs.lock.Lock()
	defer rs.lock.Unlock()

	reservation, ok := rs.reservations[id]
	if !ok {
		return structs.Reservation{}, core.NewNotFound("reservation not found")
	}

	reservation.Status = structs.ReservationStatusConfirmed

	if err := rs.store.PutReservation(reservation); err != nil {
		return structs.Reservation{}, err
	}

	rs.reservations[id] = reservation

	return reservation, nil
}
