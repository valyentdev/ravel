package logd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/containerd/console"
	"github.com/valyentdev/ravel/pkg/core"
)

type Logger struct {
	// Info to identify the machine and instance where the logger is running
	InstanceID string
	Namespace  string
	MachineID  string
	// Path to listen for logs
	PtyPath string
}

// Start will create a new goroutine that will listen for logs on the ptyPath
// When a new log is received, it will send it to the exporter
func (l *Logger) Start(ctx context.Context, exporter Exporter) error {
	// Open the ptyPath
	pty, err := os.Open(l.PtyPath)
	if err != nil {
		return err
	}

	defer pty.Close()

	// Get a console from the pty
	ptyConsole, err := console.ConsoleFromFile(pty)
	if err != nil {
		return err
	}

	defer ptyConsole.Close()

	// Create a reader to read from the console
	reader := bufio.NewReaderSize(ptyConsole, 4096)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:

			line, _, err := reader.ReadLine()

			if err != nil {
				if err.Error() == "EOF" {
					fmt.Println("EOF reached")
					break
				}
			}

			log := &core.LogEntry{
				Timestamp:  time.Now().Unix(),
				InstanceId: l.InstanceID,
				Source:     "instance",
				Level:      "info",
				Message:    string(line),
			}

			// Send the log to the exporter
			err = exporter.Send(log)

			if err != nil {
				return err
			}
		}
	}
}
