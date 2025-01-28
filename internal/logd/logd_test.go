package logd_test

import (
	"context"
	"log"
	"testing"

	"github.com/containerd/console"
	"github.com/stretchr/testify/assert"
	"github.com/valyentdev/ravel/internal/logd"
	"github.com/valyentdev/ravel/pkg/core"
)

// Channel exporter will send logs to a channel, this is with only testing purposes and should not be used in production
type ChannelExporter struct {
	Channel chan *core.LogEntry
}

func NewChannelExporter() *ChannelExporter {
	return &ChannelExporter{
		Channel: make(chan *core.LogEntry),
	}
}

func (c *ChannelExporter) Send(log *core.LogEntry) error {
	c.Channel <- log
	return nil
}

type FakeMachine struct {
	instanceId string
	namespace  string
	machineId  string
	ptyPath    string
	console    console.Console
}

func NewFakeMachine(instanceId, namespace, machineId string) (*FakeMachine, error) {
	c, path, err := console.NewPty()
	if err != nil {
		return nil, err
	}

	return &FakeMachine{
		instanceId: instanceId,
		namespace:  namespace,
		machineId:  machineId,
		ptyPath:    path,
		console:    c,
	}, nil
}

func (f *FakeMachine) GetLogger() *logd.Logger {
	return &logd.Logger{
		InstanceID: f.instanceId,
		Namespace:  f.namespace,
		MachineID:  f.machineId,
		PtyPath:    f.ptyPath,
	}
}

func (f *FakeMachine) Close() {
	f.console.Close()
}

func (f *FakeMachine) Write(data []byte) {
	f.console.Write(data)
}

func TestWatcher(t *testing.T) {
	l := log.New(log.Writer(), "", log.LstdFlags)

	l.Println("Start test")
	// Create a new channel exporter
	exporter := NewChannelExporter()

	l.Println("Exporter created")

	// Create a new LogWatcherService with the exporter
	logWatcherService := logd.LogManager{
		Loggers:  make(map[string]*logd.Logger),
		Queue:    make(chan *logd.Logger),
		Exporter: exporter,
	}
	l.Println("LogWatcherService created")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the log watcher service
	go logWatcherService.Start(ctx)

	l.Println("LogWatcherService started")
	// Create a new pty and register it with the log watcher service

	machine1, err := NewFakeMachine("instance_id_1", "default", "machine_id_1")
	if err != nil {
		t.Fatal(err)
	}
	defer machine1.Close()
	l.Println("Machine 1 created")

	machine2, err := NewFakeMachine("instance_id_2", "default", "machine_id_2")
	if err != nil {
		t.Fatal(err)
	}
	defer machine2.Close()
	l.Println("Machine 2 created")

	logWatcherService.RegisterLogger(machine1.GetLogger())
	logWatcherService.RegisterLogger(machine2.GetLogger())

	l.Println("Loggers registered")

	// Write to the first machine and check if the log is received by the exporter
	logText := "Test number 1!"

	machine1.Write([]byte(logText + "\n"))

	logEntry := <-exporter.Channel

	assert.Equal(t, logText, logEntry.Message)
	assert.Equal(t, "instance_id_1", logEntry.InstanceId)

	// Write to the second machine and check if the log is received by the exporter
	logText = "Test number 2!"

	machine2.Write([]byte(logText + ".\n"))

	logEntry = <-exporter.Channel

	assert.Equal(t, logText, logEntry.Message)
	assert.Equal(t, "instance_id_2", logEntry.InstanceId)
}
