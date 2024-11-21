package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/valyentdev/ravel/api"
)

func streamLogs(ctx huma.Context, replay []*api.LogEntry, logsChan <-chan *api.LogEntry) {
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
