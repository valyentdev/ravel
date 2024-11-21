package agentclient

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/errdefs"
)

func (a *AgentClient) getLogsRaw(ctx context.Context, path string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.baseUrl+path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errdefs.NewUnknown("failed to get logs")
	}

	return resp.Body, nil
}

func subscribeToLogs(body io.ReadCloser) ([]*api.LogEntry, <-chan *api.LogEntry, error) {
	var err error
	defer func() {
		if err != nil {
			body.Close()
		}
	}()

	logs := make(chan *api.LogEntry)
	reader := bufio.NewReaderSize(body, 8192)

	var replay []*api.LogEntry
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, nil, fmt.Errorf("failed to read logs: %w", err)
		}

		if string(line) == "null" {
			break
		}

		var log api.LogEntry

		if err := json.Unmarshal(line, &log); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal log: %w", err)
		}

		replay = append(replay, &log)
	}

	go func() {
		defer close(logs)
		defer body.Close()
		for {
			var log api.LogEntry
			line, _, err := reader.ReadLine()
			if err != nil {
				if err == io.EOF {
					return
				}
				return
			}

			if err := json.Unmarshal(line, &log); err != nil {
				return
			}
			logs <- &log
		}
	}()

	return replay, logs, nil
}
