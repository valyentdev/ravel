package logd

import (
	"context"

	"github.com/pkg/errors"
)

type LogManager struct {
	Loggers map[string]*Logger

	Queue chan *Logger

	Exporter Exporter
}

func NewLogWatcherService(config *LogdConfig) (*LogManager, error) {
	exporter, err := NewExporter(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create exporter")
	}

	return &LogManager{
		Loggers:  make(map[string]*Logger),
		Queue:    make(chan *Logger),
		Exporter: exporter,
	}, nil
}

// Start service will create a goroutine that's going to listen for a new logger on the loggers channel
// When a new logger is received, it will:
// - Check if the logger is already running
// - If the max number of loggers is reached, it will wait till a logger is removed from the loggers map
// - If not, it will wait till it can add a new goroutine to the loggers map
// - Start the logger
// - Add the logger to the loggers map
func (m *LogManager) Start(ctx context.Context) {

	for {
		select {
		case <-ctx.Done():
			return
		case logger := <-m.Queue:
			if _, ok := m.Loggers[logger.InstanceID]; ok {
				continue
			}

			m.Loggers[logger.InstanceID] = logger

			go logger.Start(ctx, m.Exporter)
		}

	}
}

func (m *LogManager) RegisterLogger(logger *Logger) {
	m.Queue <- logger
}
