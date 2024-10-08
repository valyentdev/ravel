package logging

import (
	"bufio"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/containerd/console"
	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/helper/broadcaster"
)

type LogSubscriber = *broadcaster.Subscriber[[]*core.LogEntry]

// For now, we keep logs in memory but this is not suitable for production
type InstanceLogger struct {
	instanceId string

	maxEntries int
	truncateBy int

	stop chan struct{}

	log   []*core.LogEntry
	bc    *broadcaster.Broadcaster[[]*core.LogEntry]
	mutex sync.RWMutex
}

func NewInstanceLogger(instanceId string) *InstanceLogger {
	il := &InstanceLogger{
		maxEntries: 100,
		truncateBy: 20,
		stop:       make(chan struct{}),
		instanceId: instanceId,
		log:        []*core.LogEntry{},
	}

	bc := broadcaster.NewBroadcaster[[]*core.LogEntry](broadcaster.BroadcasterOpts[[]*core.LogEntry]{
		SubsBufferSize: 5,
		GetReplay: func() [][]*core.LogEntry {
			return [][]*core.LogEntry{il.GetLog()}
		},
	})

	il.bc = bc

	return il
}

func (m *InstanceLogger) Start(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	m.stop = make(chan struct{})

	pty, err := console.ConsoleFromFile(file)
	if err != nil {
		return err
	}

	go func() {
		m.bc.Start()
		m.startReadingPty(pty)
		file.Close()
		pty.Close()
		m.bc.Stop()
	}()

	return nil
}

func (m *InstanceLogger) startReadingPty(pty console.Console) {
	for {
		select {
		case <-m.stop:
			return
		default:
			reader := bufio.NewReaderSize(pty, 4096)

			line, _, err := reader.ReadLine()
			if err != nil {
				if err == io.EOF {
					m.Stop()
					return
				}
				slog.Warn("failed to read from pty", "error", err)
				return
			}

			m.mutex.Lock()
			log := &core.LogEntry{
				Timestamp:  time.Now().Unix(),
				InstanceId: m.instanceId,
				Source:     "instance",
				Level:      "info",
				Message:    string(line),
			}
			m.appendOrResize(log)
			m.mutex.Unlock()
			m.bc.Publish([]*core.LogEntry{log})
		}
	}
}

func (m *InstanceLogger) appendOrResize(log *core.LogEntry) {
	if len(m.log) < m.maxEntries {
		m.log = append(m.log, log)
		return
	}

	m.log = append(m.log[m.truncateBy:], log)
}

func (m *InstanceLogger) Stop() {
	close(m.stop)
}

func (m *InstanceLogger) GetLog() []*core.LogEntry {
	m.mutex.RLock()
	logs := make([]*core.LogEntry, len(m.log))
	copy(logs, m.log)
	m.mutex.RUnlock()
	return logs
}

func (m *InstanceLogger) Subscribe() LogSubscriber {
	return m.bc.Subscribe()
}
