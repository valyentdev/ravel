package store

import (
	"encoding/json"

	"github.com/valyentdev/ravel/agent/structs"
	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/api/errdefs"
	"go.etcd.io/bbolt"
)

const (
	machineInstanceMachineKey = "machine"
	machineInstanceVersionKey = "version"
	machineInstanceStateKey   = "state"
	machineInstanceNetworkKey = "network"
)

func assertMachineInstancesBucketExists(bucket *bbolt.Bucket) {
	if bucket == nil {
		panic("instances bucket not found the Init function should have been called")
	}
}

func (s *Store) LoadMachineInstances() ([]structs.MachineInstance, error) {
	tx, err := s.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	machineInstances := tx.Bucket(machineInstancesBucket)
	assertMachineInstancesBucketExists(machineInstances)

	var instances []structs.MachineInstance

	err = machineInstances.ForEach(func(k, v []byte) error {
		machine := machineInstances.Bucket(k)
		if machine == nil {
			return nil
		}

		m := machine.Get([]byte(machineInstanceMachineKey))
		mv := machine.Get([]byte(machineInstanceVersionKey))
		ms := machine.Get([]byte(machineInstanceStateKey))
		mn := machine.Get([]byte(machineInstanceNetworkKey))

		var mi structs.MachineInstance
		if err := json.Unmarshal(m, &mi.Machine); err != nil {
			return err
		}

		if err := json.Unmarshal(mv, &mi.Version); err != nil {
			return err
		}

		if err := json.Unmarshal(ms, &mi.State); err != nil {
			return err
		}

		if err := json.Unmarshal(mn, &mi.Network); err != nil {
			return err
		}

		instances = append(instances, mi)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return instances, nil
}

func (s *Store) CreateMachineInstance(mi structs.MachineInstance) error {
	tx, err := s.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	machineInstances := tx.Bucket(machineInstancesBucket)
	assertMachineInstancesBucketExists(machineInstances)

	machine, err := machineInstances.CreateBucket([]byte(mi.Machine.Id))
	if err != nil {
		return err
	}

	m, err := json.Marshal(mi.Machine)
	if err != nil {
		return err
	}

	mv, err := json.Marshal(mi.Version)
	if err != nil {
		return err
	}

	ms, err := json.Marshal(mi.State)
	if err != nil {
		return err
	}

	mn, err := json.Marshal(mi.Network)
	if err != nil {
		return err
	}

	if err = machine.Put([]byte(machineInstanceMachineKey), m); err != nil {
		return err
	}

	if err = machine.Put([]byte(machineInstanceVersionKey), mv); err != nil {
		return err
	}

	if err = machine.Put([]byte(machineInstanceStateKey), ms); err != nil {
		return err
	}

	if err = machine.Put([]byte(machineInstanceNetworkKey), mn); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) UpdateMachineInstance(id string, mi *structs.MachineInstanceState, me *api.MachineEvent) error {
	tx, err := s.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	machineInstances := tx.Bucket(machineInstancesBucket)
	assertMachineInstancesBucketExists(machineInstances)

	machine := machineInstances.Bucket([]byte(id))
	if machine == nil {
		return errdefs.NewNotFound("machine not found, skipping update")
	}

	ms, err := json.Marshal(mi)
	if err != nil {
		return err
	}

	if err = machine.Put([]byte(machineInstanceStateKey), ms); err != nil {
		return err
	}

	if me != nil {
		machineEvents := tx.Bucket(eventsBucket)
		assertEventsBucketExists(machineEvents)

		meBytes, err := json.Marshal(me)
		if err != nil {
			return err
		}

		if err = machineEvents.Put([]byte(me.Id), meBytes); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) DeleteMachineInstance(id string) error {
	tx, err := s.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	machineInstances := tx.Bucket(machineInstancesBucket)
	assertMachineInstancesBucketExists(machineInstances)

	if err = machineInstances.DeleteBucket([]byte(id)); err != nil {
		return err
	}

	return tx.Commit()
}
