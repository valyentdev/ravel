package streamutil

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/valyentdev/ravel/api"
)

func SubscribeToLogs(body io.ReadCloser) ([]*api.LogEntry, <-chan *api.LogEntry, error) {
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

func StreamLogs(ctx huma.Context, replay []*api.LogEntry, logsChan <-chan *api.LogEntry) {
	ctx.SetHeader("Content-Type", "application/x-ndjson")
	ctx.SetStatus(http.StatusOK)

	bw := ctx.BodyWriter()
	rw := bw.(http.ResponseWriter)
	rc := http.NewResponseController(rw)

	err := rc.SetWriteDeadline(time.Now().Add(30 * time.Second))
	if err != nil {
		slog.Error("Failed to set write deadline", "error", err)
		return
	}
	err = rc.Flush()
	if err != nil {
		slog.Error("Failed to flush response", "error", err)
		return
	}

	for _, log := range replay {
		bytes, err := json.Marshal(log)
		if err != nil {
			return
		}

		_, err = rw.Write(bytes)
		if err != nil {
			return
		}
		_, err = rw.Write([]byte("\n"))
		if err != nil {
			return
		}
	}

	rw.Write([]byte("null\n")) // End of replay
	err = rc.Flush()
	if err != nil {
		slog.Error("Failed to flush response", "error", err)
		return
	}

	for log := range logsChan {
		bytes, err := json.Marshal(log)
		if err != nil {
			return
		}

		_, err = rw.Write(bytes)
		if err != nil {
			return
		}
		_, err = rw.Write([]byte("\n"))

		if err != nil {
			return
		}

		err = rc.Flush()
		if err != nil {
			return
		}
	}
}
