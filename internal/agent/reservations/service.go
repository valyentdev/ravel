package reservations

import (
	"context"
	"log/slog"
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
	lock                 sync.Mutex
	store                *store.Store
	max                  core.Resources
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
	}
}

func (rs *ReservationService) Init(ctx context.Context) error {
	reservations, err := rs.store.ListReservations(ctx)
	if err != nil {
		return err
	}

	for _, res := range reservations {
		subnet := res.LocalIPV4Subnet.LocalConfig().Network

		if err := rs.localSubnetAllocator.Allocate(subnet); err != nil {
			return err
		}
	}

	return nil
}

func (rs *ReservationService) CreateReservation(ctx context.Context, id string, res core.Resources, status structs.ReservationStatus) (reservation structs.Reservation, before core.Resources, after core.Resources, err error) {
	rs.lock.Lock()
	defer rs.lock.Unlock()

	localIPV4Subnet, err := rs.localSubnetAllocator.AllocateNext()
	if err != nil {
		return
	}

	slog.Info("Allocated local subnet", "subnet", localIPV4Subnet)

	defer func() {
		if err != nil {
			rs.localSubnetAllocator.Release(&localIPV4Subnet)
		}
	}()

	alloc := structs.Reservation{
		Id:              id,
		Cpus:            res.Cpus,
		Memory:          res.Memory,
		LocalIPV4Subnet: networking.LocalIPV4Subnet(localIPV4Subnet.String()),
		Status:          status,
		CreatedAt:       time.Now(),
	}

	before, err = rs.store.GetReservedResources(ctx)
	if err != nil {
		return
	}

	after = core.Resources{
		Cpus:   before.Cpus + res.Cpus,
		Memory: before.Memory + res.Memory,
	}

	if after.Cpus > rs.max.Cpus || after.Memory > rs.max.Memory {
		err = ErrNotEnoughResources
		return
	}

	if err = rs.store.CreateReservation(ctx, alloc); err != nil {
		return
	}

	return
}

func (rs *ReservationService) GetReservation(ctx context.Context, id string) (structs.Reservation, error) {
	return rs.store.GetReservation(ctx, id)
}

func (rs *ReservationService) DeleteReservation(ctx context.Context, id string) error {
	rs.lock.Lock()
	defer rs.lock.Unlock()

	reservation, err := rs.store.GetReservation(ctx, id)
	if err != nil {
		if core.IsNotFound(err) {
			return nil
		}
		return err
	}

	subnet := reservation.LocalIPV4Subnet.LocalConfig().Network

	if err := rs.localSubnetAllocator.Release(subnet); err != nil {
		return err
	}

	if err := rs.store.DeleteReservation(ctx, id); err != nil {
		return err
	}

	return nil
}

func (rs *ReservationService) ListReservations(ctx context.Context) ([]structs.Reservation, error) {
	return rs.store.ListReservations(ctx)
}

func (rs *ReservationService) ConfirmReservation(ctx context.Context, id string) (structs.Reservation, error) {
	rs.lock.Lock()
	defer rs.lock.Unlock()

	reservation, err := rs.store.GetReservation(ctx, id)
	if err != nil {
		return structs.Reservation{}, err
	}

	if reservation.Status == structs.ReservationStatusConfirmed {
		return reservation, nil
	}

	reservation.Status = structs.ReservationStatusConfirmed

	if err := rs.store.UpdateReservation(ctx, reservation); err != nil {
		return reservation, err
	}

	return reservation, nil
}
