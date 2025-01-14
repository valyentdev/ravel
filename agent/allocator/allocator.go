package allocator

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/valyentdev/ravel/agent/structs"
	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/api/errdefs"
)

type AllocationsStore interface {
	LoadAllocations() ([]structs.Allocation, error)
	PutAllocation(structs.Allocation) error
	DeleteAllocation(string) error
}

var ErrNotEnoughResources = errdefs.NewResourcesExhausted("not enough resources")

type Allocator struct {
	store        AllocationsStore
	max          api.Resources
	current      api.Resources
	lock         sync.RWMutex
	reservations map[string]structs.Allocation
}

func (a *Allocator) Max() api.Resources {
	return a.max
}

func New(store AllocationsStore, totalResources api.Resources) (*Allocator, error) {
	a := &Allocator{
		store:        store,
		max:          totalResources,
		reservations: make(map[string]structs.Allocation),
	}
	reservations, err := a.store.LoadAllocations()
	if err != nil {
		return nil, err
	}

	for _, res := range reservations {
		a.reservations[res.Id] = res
		if res.Status == structs.AllocationStatusConfirmed {
			a.current = a.current.Add(res.Resources)
		}
	}

	return a, nil
}

func (a *Allocator) CreateAllocation(id string, res api.Resources) (reservation structs.Allocation, before api.Resources, after api.Resources, err error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	alloc := structs.Allocation{
		Id:        id,
		Resources: res,
		Status:    structs.AllocationStatusDangling,
		CreatedAt: time.Now(),
	}

	before = a.current

	after = before.Add(res)

	if after.GT(a.max) {
		err = ErrNotEnoughResources
		return
	}

	a.current = after

	a.reservations[id] = alloc

	go a.gc(id)

	return
}

func (a *Allocator) GetAllocation(id string) (structs.Allocation, error) {
	a.lock.RLock()
	defer a.lock.RUnlock()

	reservation, ok := a.reservations[id]
	if !ok {
		return structs.Allocation{}, errdefs.NewNotFound("reservation not found")
	}

	return reservation, nil
}

func (a *Allocator) DeleteAllocation(id string) error {
	a.lock.Lock()
	defer a.lock.Unlock()
	reservation, ok := a.reservations[id]
	if !ok {
		slog.Warn("reservation not found", "id", id)
		return nil
	}

	a.current = a.current.Sub(reservation.Resources)

	if err := a.store.DeleteAllocation(id); err != nil {
		return err
	}

	delete(a.reservations, id)

	return nil
}

func (a *Allocator) ListAllocations(ctx context.Context) []structs.Allocation {
	a.lock.RLock()

	reservations := make([]structs.Allocation, len(a.reservations))
	for _, res := range a.reservations {
		reservations = append(reservations, res)
	}
	a.lock.RUnlock()
	return reservations
}

func (a *Allocator) ConfirmAllocation(id string) (structs.Allocation, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	reservation, ok := a.reservations[id]
	if !ok {
		return structs.Allocation{}, errdefs.NewNotFound("reservation not found")
	}

	reservation.Status = structs.AllocationStatusConfirmed

	if err := a.store.PutAllocation(reservation); err != nil {
		return structs.Allocation{}, err
	}

	a.reservations[id] = reservation

	return reservation, nil
}
