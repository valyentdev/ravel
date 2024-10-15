package store

import "github.com/valyentdev/ravel/pkg/core"

/**

allocations/
	{alloc-id}/
		instance -> core.Instance
		events/
			{event-id} -> core.InstanceEvent


**/

type StoreI interface {
	LoadInstances() ([]*core.Instance, error)
}

type TxI interface {
	PutInstance(instance *core.Instance) error
	PutInstanceEvent(event *core.InstanceEvent) error
}
