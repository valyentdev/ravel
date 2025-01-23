package agent

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/valyentdev/ravel/agent/allocator"
	"github.com/valyentdev/ravel/agent/machinerunner"
	"github.com/valyentdev/ravel/agent/machinerunner/state"
	"github.com/valyentdev/ravel/agent/node"
	"github.com/valyentdev/ravel/agent/server"
	"github.com/valyentdev/ravel/core/cluster"
	"github.com/valyentdev/ravel/core/cluster/corrosion"
	"github.com/valyentdev/ravel/core/cluster/placement"
	"github.com/valyentdev/ravel/core/config"
	"github.com/valyentdev/ravel/core/daemon/network"
	"github.com/valyentdev/ravel/internal/eventer"
	"github.com/valyentdev/ravel/internal/mtls"

	"github.com/valyentdev/ravel/api"

	"github.com/valyentdev/ravel/runtime"
)

type Agent struct {
	node      *node.Node
	config    *config.AgentConfig
	nc        *nats.Conn
	cluster   cluster.ClusterState
	store     Store
	machines  machinerunner.Store
	eventer   *eventer.Eventer[api.MachineEvent]
	allocator *allocator.Allocator
	runtime   *runtime.Runtime
	server    *server.AgentServer
	placement *placement.Listener
	network   *network.NetworkService
}

type Config struct {
	Agent     *config.AgentConfig
	Nats      *config.NatsConfig
	Corrosion *config.CorrosionConfig
}

type Store interface {
	state.Store
	allocator.AllocationsStore
}

func New(config Config, store Store, runtime *runtime.Runtime, netservice *network.NetworkService) (*Agent, error) {
	slog.Info("Initializing agent", "node_id", config.Agent.NodeId, "address", config.Agent.Address)

	nc, err := config.Nats.Connect()
	if err != nil {
		return nil, err
	}

	cs, err := corrosion.New(config.Corrosion.PgWireAddr)
	if err != nil {
		return nil, err
	}

	slog.Info("Initializing allocator")
	allocator, err := allocator.New(store, config.Agent.Resources)
	if err != nil {
		return nil, fmt.Errorf("failed to create reservation service: %w", err)
	}

	node := node.NewNode(cs, api.Node{
		Id:            config.Agent.NodeId,
		Address:       config.Agent.Address,
		AgentPort:     config.Agent.Port,
		HttpProxyPort: config.Agent.HttpProxyPort,
		Region:        config.Agent.Region,
		HeartbeatedAt: time.Now(),
	})

	agent := &Agent{
		node:      node,
		config:    config.Agent,
		nc:        nc,
		cluster:   cs,
		store:     store,
		machines:  machinerunner.NewStore(),
		eventer:   newMachineEventer(store, nc),
		allocator: allocator,
		runtime:   runtime,
		placement: placement.NewListener(nc),
		network:   netservice,
	}

	events, err := store.LoadMachineInstanceEvents()
	if err != nil {
		return nil, err
	}

	agent.eventer.Start(events)

	return agent, nil
}

func (a *Agent) startListener() (net.Listener, error) {
	laddr := fmt.Sprintf("%s:%d", a.config.Address, a.config.Port)
	if a.config.TLS == nil {
		return net.Listen("tcp", laddr)
	}

	cert, err := a.config.TLS.LoadCert()
	if err != nil {
		return nil, err
	}

	ca, err := a.config.TLS.LoadCA()
	if err != nil {
		return nil, err
	}

	tlsConfig := tls.Config{
		Certificates:     []tls.Certificate{cert},
		ClientCAs:        ca,
		VerifyConnection: mtls.VerifyAgentConnection,
	}

	if !a.config.TLS.SkipVerifyClient {
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	listener, err := tls.Listen("tcp", laddr, &tlsConfig)
	if err != nil {
		return nil, err
	}

	return listener, nil
}

func (a *Agent) Start() error {
	slog.Info("Starting agent")
	machines, err := a.store.LoadMachineInstances()
	if err != nil {
		return err
	}

	for _, m := range machines {
		err := a.network.Allocate(m.Network)
		if err != nil {
			return err
		}
		machine := a.newMachine(m)
		a.machines.AddMachine(machine)
		go machine.Run()
	}

	a.server = server.NewAgentServer(a)

	listener, err := a.startListener()
	defer func() {
		if err != nil {
			listener.Close()
		}
	}()

	go a.server.Serve(listener)

	if err = a.node.Start(); err != nil {
		return err
	}

	if err = a.startPlacementHandler(); err != nil {
		return err
	}

	return nil
}

func (d *Agent) Stop(ctx context.Context) error {
	d.placement.Stop()

	return d.server.Shutdown(ctx)
}
