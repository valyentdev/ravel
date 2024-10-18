package store

import (
	"encoding/json"

	"github.com/oklog/ulid"
	"github.com/valyentdev/ravel/pkg/core"
	"go.etcd.io/bbolt"
)

func getUnreportedInstanceEvents(events *bbolt.Bucket, lastReported []byte) ([]core.InstanceEvent, error) {
	eventList := []core.InstanceEvent{}

	cursor := events.Cursor()

	cursor.Seek(lastReported)

	for k, v := cursor.Next(); k != nil; k, v = cursor.Next() {
		var event core.InstanceEvent

		if err := json.Unmarshal(v, &event); err != nil {
			return nil, err
		}

		eventList = append(eventList, event)
	}

	return eventList, nil
}

func (s *Store) SetLastReportedEventId(instanceId string, eventId ulid.ULID) error {
	tx, err := s.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	instances := tx.Bucket(instancesBucket)

	instance := instances.Bucket([]byte(instanceId))

	err = instance.Put(lastReportedEventIdKey, eventId[:])
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
