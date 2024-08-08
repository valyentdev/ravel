package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/valyentdev/ravel/pkg/core"
)

var instanceEventFields = []string{
	"id",
	"instance_id",
	"type",
	"origin",
	"payload",
	"status",
	"reported",
	"timestamp",
}

var instanceEventColumns = allColumns(instanceEventFields)
var instanceEventPlaceholders = placeholders(len(instanceEventFields))

func scanEvent(scannable scannable) (core.InstanceEvent, error) {
	event := core.InstanceEvent{}
	var payloadJSON []byte

	err := scannable.Scan(
		&event.Id,
		&event.InstanceId,
		&event.Type,
		&event.Origin,
		&payloadJSON,
		&event.Status,
		&event.Reported,
		&event.Timestamp,
	)

	if err != nil {
		return event, err
	}

	payload, err := core.UnmarshalEventPayload(event.Type, payloadJSON)

	if err != nil {
		return event, err
	}

	event.Payload = payload

	return event, nil

}

func (q *Queries) ListInstanceEvents(ctx context.Context, instanceId string) ([]core.InstanceEvent, error) {
	rows, err := q.db.QueryContext(ctx, fmt.Sprintf("SELECT %s FROM instance_events WHERE instance_id = ?1 ORDER BY timestamp DESC", instanceEventColumns), instanceId)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var events []core.InstanceEvent
	for rows.Next() {
		event, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	return events, nil
}

func (q *Queries) GetLastInstanceEvent(ctx context.Context, instanceId string) (core.InstanceEvent, error) {
	row := q.db.QueryRowContext(
		ctx,
		fmt.Sprintf(`SELECT %s FROM instance_events WHERE instance_id = ?1 ORDER BY timestamp DESC LIMIT 1`, instanceEventColumns),
		instanceId,
	)

	return scanEvent(row)
}

func (q *Queries) StoreInstanceEvent(ctx context.Context, event *core.InstanceEvent) error {

	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("INSERT INTO instance_events (%s) VALUES (%s)", instanceEventColumns, instanceEventPlaceholders)

	_, err = q.db.ExecContext(ctx, query,
		event.Id,
		event.InstanceId,
		event.Type,
		event.Origin,
		payload,
		event.Status,
		event.Reported,
		event.Timestamp,
	)

	if err != nil {
		return err
	}

	return nil
}
