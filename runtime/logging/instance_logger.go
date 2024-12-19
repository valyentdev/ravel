package logging

import (
	"bufio"
	"io"
	"os"
	"sync"
	"time"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/pkg/pubsub"
)

type LogSubscriber = pubsub.Subscriber[*api.LogEntry]

// For now, we keep logs in memory but this is not suitable for production
type InstanceLogger struct {
	instanceId string

	maxEntries int
	truncateBy int

	stop chan struct{}

	log   []*api.LogEntry
	bc    *pubsub.MessageBroadcaster[*api.LogEntry]
	mutex sync.RWMutex
}

func NewInstanceLogger(instanceId string) *InstanceLogger {
	il := &InstanceLogger{
		maxEntries: 100,
		truncateBy: 20,
		stop:       make(chan struct{}),
		instanceId: instanceId,
		log:        []*api.LogEntry{},
	}

	bc := pubsub.NewMessageBroadcaster(pubsub.BroadcasterOpts[*api.LogEntry]{
		SubsBufferSize: 5,
	})

	il.bc = bc

	return il
}

func (m *InstanceLogger) Start(path string) error {
	file, err := os.OpenFile(path, os.O_RDONLY, os.ModeNamedPipe)
	if err != nil {
		return err
	}

	m.stop = make(chan struct{})

	go func() {
		m.bc.Start()
		m.startReading(file)
		file.Close()
		m.bc.Stop()
	}()

	return nil
}

func (m *InstanceLogger) startReading(r io.Reader) {
	for {
		select {
		case <-m.stop:
			return
		default:
			reader := bufio.NewReaderSize(r, 4096)

			line, _, err := reader.ReadLine()
			if err != nil {
				return
			}

			log := &api.LogEntry{
				Timestamp:  time.Now().Unix(),
				InstanceId: m.instanceId,
				Source:     "instance",
				Level:      "info",
				Message:    string(line),
			}
			m.mutex.Lock()
			m.appendOrResize(log)
			m.bc.Publish(log)
			m.mutex.Unlock()

		}
	}
}

func (m *InstanceLogger) appendOrResize(log *api.LogEntry) {
	if len(m.log) < m.maxEntries {
		m.log = append(m.log, log)
		return
	}

	m.log = append(m.log[m.truncateBy:], log)
}

func (m *InstanceLogger) Stop() {
	close(m.stop)
}

func (m *InstanceLogger) getLog() []*api.LogEntry {
	logs := make([]*api.LogEntry, len(m.log))
	copy(logs, m.log)
	return logs
}

func (m *InstanceLogger) GetLog() []*api.LogEntry {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.getLog()
}

func (m *InstanceLogger) Subscribe() ([]*api.LogEntry, *LogSubscriber) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	log := m.getLog()
	return log, m.bc.Subscribe()
}
