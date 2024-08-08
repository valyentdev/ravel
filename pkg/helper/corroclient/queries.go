package corroclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
)

var ErrNoRows = errors.New("corroclient: no rows")

func (c *CorroClient) QueryContext(ctx context.Context, stmt Statement) (*Rows, error) {
	payload, err := json.Marshal(stmt)
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBuffer(payload)

	request, err := http.NewRequest("POST", c.getURL("/v1/queries"), buffer)
	if err != nil {
		return nil, err
	}

	resp, err := c.request(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("corroclient: invalid status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	reader := bufio.NewReader(resp.Body)

	var columns Columns

	rows := []*Row{}

	for {
		data, err := reader.ReadBytes('\n')
		if err != nil {
			return nil, err
		}

		var e event
		if err := json.Unmarshal(data, &e); err != nil {
			return nil, err
		}

		if e.Columns != nil {
			columns = e.Columns
			continue
		}

		if e.Row != nil {
			row, err := readRow(e.Row)
			if err != nil {
				return nil, err
			}

			rows = append(rows, row)
		}

		if e.EOQ != nil {
			break
		}

	}

	if len(rows) == 0 {
		return nil, ErrNoRows
	}

	return &Rows{
		columns:      columns,
		rows:         rows,
		mutex:        sync.RWMutex{},
		currentIndex: -1,
	}, nil
}

func (c *CorroClient) QueryRowContext(ctx context.Context, stmt Statement) (*Row, error) {
	rows, err := c.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return nil, ErrNoRows // should never append but just in case...
	}

	row := rows.rows[rows.currentIndex]
	row.columns = rows.columns

	return row, nil
}
