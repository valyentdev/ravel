package corroclient

import (
	"errors"
)

type Statement struct {
	Query       string         `json:"query"`
	Params      []any          `json:"params,omitempty"`
	NamedParams map[string]any `json:"named_params,omitempty"`
}

type EventType string

const (
	EventTypeRow     EventType = "row"
	EventTypeEOQ     EventType = "eoq"
	EventTypeChange  EventType = "change"
	EventTypeColumns EventType = "columns"
	EventTypeError   EventType = "error"
)

type Event interface {
	Type() EventType
}

type ChangeType string

const (
	ChangeTypeInsert ChangeType = "insert"
	ChangeTypeUpdate ChangeType = "update"
	ChangeTypeDelete ChangeType = "delete"
)

type Change struct {
	ChangeId   int64      `json:"change_id"`
	ChangeType ChangeType `json:"change_type"`
	Row        *Row       `json:"row"`
}

func (c *Change) Type() EventType {
	return EventTypeChange
}

type Row struct {
	rowId   int64
	values  []any
	columns []string
}

func (r *Row) RowId() int64 {
	return r.rowId
}

func (r *Row) Type() EventType {
	return EventTypeRow
}

type EOQ struct {
	ChangeId int64   `json:"change_id"`
	Time     float64 `json:"time"`
}

func (e *EOQ) Type() EventType {
	return EventTypeEOQ
}

type Columns []string

func (c Columns) Type() EventType {
	return EventTypeColumns
}

type Error struct {
	Message string `json:"message"`
}

func (e *Error) Type() EventType {
	return EventTypeError
}

var ErrInvalidRow = errors.New("corrosubs: Invalid row")

func readRow(data []any) (*Row, error) {
	if len(data) != 2 {
		return nil, ErrInvalidRow
	}

	rowIdFloat, ok := data[0].(float64)
	if !ok {
		return nil, ErrInvalidRow
	}

	rowId := int64(rowIdFloat)
	values, ok := data[1].([]any)
	if !ok {
		return nil, ErrInvalidRow
	}

	return &Row{
		rowId:  rowId,
		values: values,
	}, nil
}

var ErrInvalidChange = errors.New("corrosubs: Invalid change")

func readChange(data []any) (*Change, error) {
	if len(data) != 4 {
		return nil, ErrInvalidRow
	}

	changeType, ok := data[0].(string)
	if !ok {
		return nil, ErrInvalidChange
	}

	rowId, ok := data[1].(float64)
	if !ok {
		return nil, ErrInvalidChange
	}

	values, ok := data[2].([]any)
	if !ok {
		return nil, ErrInvalidChange
	}

	changeId, ok := data[3].(float64)
	if !ok {
		return nil, ErrInvalidChange
	}

	return &Change{
		ChangeId:   int64(changeId),
		ChangeType: ChangeType(changeType),
		Row: &Row{
			rowId:  int64(rowId),
			values: values,
		},
	}, nil
}
