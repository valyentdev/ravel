package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/valyentdev/ravel/agent/allocator"
	"github.com/valyentdev/ravel/agent/machine"
	"github.com/valyentdev/ravel/agent/node"
	"github.com/valyentdev/ravel/agent/store"
	"github.com/valyentdev/ravel/core/cluster"
	"github.com/valyentdev/ravel/core/config"
	"github.com/valyentdev/ravel/core/corrosion"
	"github.com/valyentdev/ravel/core/instance"

	"github.com/valyentdev/ravel/agent/structs"
	"github.com/valyentdev/ravel/api"

	"github.com/valyentdev/ravel/runtime"
)

type Agent struct {
	node          *node.Node
	config        config.AgentConfig
	nc            *nats.Conn
	cluster       cluster.ClusterState
	store         *store.Store
	machines      machine.Store
	eventer       *eventer
	stateReporter *stateReporter
	allocator     *allocator.Allocator
	runtime       *runtime.Runtime
}

var _ structs.Agent = (*Agent)(nil)

func New(config config.RavelConfig) (*Agent, error) {
	slog.Info("Initializing agent", "node_id", config.Agent.NodeId, "address", config.Agent.Address)
	store, err := store.NewStore(config.Agent.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create store: %w", err)
	}

	err = store.Init()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize store: %w", err)
	}

	natsOptions := []nats.Option{}
	if config.Nats.CredFile != "" {
		natsOptions = append(natsOptions, nats.UserCredentials(config.Nats.CredFile, config.Nats.CredFile))
		natsOptions = append(natsOptions, nats.MaxReconnects(-1))
	}

	slog.Info("Initializing nats")
	nc, err := nats.Connect(config.Nats.Url, natsOptions...)
	if err != nil {
		return nil, err
	}

	allocator, err := allocator.New(store, config.Agent.Resources.Resources())
	if err != nil {
		return nil, fmt.Errorf("failed to create reservation service: %w", err)
	}

	if err := allocator.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize reservation service: %w", err)
	}

	cs, err := corrosion.Connect(config.Corrosion.Config())
	if err != nil {
		return nil, err
	}

	sr := newStateReporter(cs)

	node := node.NewNode(cs, api.Node{
		Id:            config.Agent.NodeId,
		Address:       config.Agent.Address,
		AgentPort:     config.Agent.AgentPort,
		HttpProxyPort: config.Agent.HttpProxyPort,
		Region:        config.Agent.Region,
		HeartbeatedAt: time.Now(),
	})

	agent := &Agent{
		node:          node,
		config:        config.Agent,
		nc:            nc,
		cluster:       cs,
		store:         store,
		machines:      machine.NewStore(),
		stateReporter: sr,
		allocator:     allocator,
	}

	eventer := newEventer(store, agent.reportEvent)

	agent.eventer = eventer

	er := newEventReporter(agent)

	runtime, err := runtime.New(
		store,
		er,
		config.Agent.InitBinary,
		config.Agent.LinuxKernel,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create runtime: %w", err)
	}

	agent.runtime = runtime

	return agent, nil
}

func (d *Agent) Start(ctx context.Context) error {
	machines, err := d.store.LoadMachineInstances()
	if err != nil {
		return err
	}

	for _, m := range machines {
		machine := d.newMachine(m)
		d.machines.AddMachine(machine)
	}

	if err := d.node.Start(); err != nil {
		return err
	}

	if err := d.runtime.Start(); err != nil {
		return err
	}

	if err := d.startPlacementHandler(); err != nil {
		return err
	}

	go d.eventer.Start()

	d.machines.Foreach(func(m *machine.Machine) {
		go m.Recover()
	})

	return nil
}

func (d *Agent) Stop() error {
	return d.store.Close()
}

type eventReporter struct {
	agent *Agent
}

func newEventReporter(agent *Agent) *eventReporter {
	return &eventReporter{
		agent: agent,
	}
}

func (a *Agent) reportEvent(event api.MachineEvent) error {
	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = a.nc.Request("machines.events", bytes, time.Second)
	if err != nil {
		return err
	}

	return nil

}

func (e *eventReporter) ReportInstanceEvent(event instance.Event) {
	if event.InstanceMetadata.MachineId == "" {
		return
	}

	machine, err := e.agent.machines.GetMachine(event.InstanceMetadata.MachineId)
	if err != nil {
		slog.Error("Failed to get machine", "error", err)
		return
	}

	machine.ProcessInstanceEvent(event)
}

var _ instance.EventReporter = (*eventReporter)(nil)
