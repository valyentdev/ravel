package corroclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func (c *CorroClient) ExecContext(ctx context.Context, stmts []Statement) (*ExecResult, error) {
	payload, err := json.Marshal(stmts)
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBuffer(payload)

	request, err := http.NewRequest("POST", c.getURL("/v1/transactions"), buffer)
	if err != nil {
		return nil, err
	}

	resp, err := c.request(request)
	if err != nil {
		return nil, err

	}

	if resp.StatusCode != http.StatusOK {
		bodyErr, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("corroclient: invalid status code: %d, body: %s", resp.StatusCode, string(bodyErr))
	}

	var execResult ExecResult
	err = json.NewDecoder(resp.Body).Decode(&execResult)
	if err != nil {
		return nil, err
	}

	return &execResult, nil
}

type ExecResult struct {
	Results []Result `json:"results"`
}

type Result struct {
	Error       string  `json:"error"`
	RowAffected int     `json:"rows_affected"`
	Time        float64 `json:"time"`
}

func (r *Result) Err() error {
	if r.Error != "" {
		return errors.New(r.Error)
	}
	return nil
}
