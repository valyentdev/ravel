package logging

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/containerd/console"
)

type LogSubscriber struct {
	ctx context.Context
	ch  chan []byte
}

type InstanceLogger struct {
	bufSize    int
	maxLogSize int
	truncateBy int

	pts         *os.File
	log         []byte
	subscribers []LogSubscriber
}

func NewMachineLogger(pts string) (*InstanceLogger, error) {
	ptsFile, err := os.Open(pts)
	if err != nil {
		return nil, err
	}

	return &InstanceLogger{
		pts:         ptsFile,
		bufSize:     256,
		maxLogSize:  10 * 1024, // 10KB
		truncateBy:  2 * 1024,  // 2KB
		subscribers: []LogSubscriber{},
	}, nil
}

func (m *InstanceLogger) write(buf []byte) {
	if len(m.log) > m.maxLogSize {
		slog.Debug("truncating logs")
		m.log = m.log[m.truncateBy:]
	}

	m.log = append(m.log, buf...)
}

func (m *InstanceLogger) broadcast(buf []byte) {
	for i, sub := range m.subscribers {
		select {
		case <-sub.ctx.Done():
			slog.Debug("removing subscriber")
			close(sub.ch)
			m.subscribers = append(m.subscribers[:i], m.subscribers[i+1:]...)
		default:
			sub.ch <- buf
		}
	}
}

func (m *InstanceLogger) Start(ctx context.Context) error {
	defer m.pts.Close()

	pty, err := console.ConsoleFromFile(m.pts)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			buf := make([]byte, 1024)
			n, err := pty.Read(buf)
			fmt.Print(string(buf[:n]))
			if err != nil {
				return err
			}
			m.write(buf[:n])
			m.broadcast(buf[:n])

		}
	}

}

func (m *InstanceLogger) Subscribe(ctx context.Context, ch chan []byte) {
	sub := LogSubscriber{ctx: ctx, ch: ch}
	m.subscribers = append(m.subscribers, sub)
	ch <- m.log

	<-ctx.Done()
	for i, sub := range m.subscribers {
		if sub.ctx == ctx {
			m.subscribers = append(m.subscribers[:i], m.subscribers[i+1:]...)
			close(sub.ch)
			return
		}
	}

}

func (m *InstanceLogger) GetLog() []byte {
	return m.log
}
