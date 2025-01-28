package logd

import (
	"fmt"
	"time"

	"github.com/valyentdev/ravel/pkg/core"
)

type Exporter interface {
	Send(*core.LogEntry) error
}

type StdoutExporter struct{}

func (s *StdoutExporter) Send(log *core.LogEntry) error {
	_, err := fmt.Printf("%s %s %s %s %s\n", time.Unix(log.Timestamp, 0).Format(time.RFC3339), log.InstanceId, log.Source, log.Level, log.Message)
	return err
}

func NewExporter(config *LogdConfig) (Exporter, error) {
	switch config.Exporter {
	case "stdout":
		return &StdoutExporter{}, nil
	default:
		return nil, fmt.Errorf("unknown exporter: %s", config.Exporter)
	}
}
