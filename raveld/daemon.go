package raveld

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/valyentdev/ravel/agent"
	"github.com/valyentdev/ravel/core/config"
	"github.com/valyentdev/ravel/core/daemon"
	"github.com/valyentdev/ravel/core/daemon/network"
	"github.com/valyentdev/ravel/raveld/server"
	"github.com/valyentdev/ravel/raveld/store"
	"github.com/valyentdev/ravel/runtime"
)

type Daemon struct {
	runtime *runtime.Runtime
	store   *store.Store
	config  *config.RavelConfig
	network *network.NetworkService

	server *server.DaemonServer
	agent  *agent.Agent
}

var _ daemon.Daemon = (*Daemon)(nil)

func NewDaemon(config config.RavelConfig) (*Daemon, error) {
	daemonConfig := config.Daemon

	store, err := store.NewStore(daemonConfig.DatabasePath)
	if err != nil {
		return nil, err
	}

	err = store.Init()
	if err != nil {
		return nil, err
	}

	runtime, err := runtime.New(config.Daemon.Runtime, config.Registries, store)
	if err != nil {
		return nil, err
	}

	networkService := network.NewNetworkService()

	daemon := &Daemon{
		runtime: runtime,
		store:   store,
		config:  &config,
		network: networkService,
	}

	daemon.server = server.NewDaemonServer(daemon)

	hasAgent := daemonConfig.Agent != nil

	if hasAgent {
		a, err := agent.New(agent.Config{
			Agent:     daemonConfig.Agent,
			Nats:      config.Nats,
			Corrosion: config.Corrosion,
		}, store, runtime, networkService)
		if err != nil {
			return nil, err
		}

		daemon.agent = a
	}

	return daemon, nil
}

func (d *Daemon) Start() error {
	err := d.runtime.Start()
	if err != nil {
		return err
	}
	instances := d.runtime.ListInstances()
	for _, instance := range instances {
		if instance.Metadata.MachineId != "" {
			err := d.network.Allocate(instance.Network)
			if err != nil {
				return err
			}
		}
	}

	if d.agent != nil {
		err = d.agent.Start()
		if err != nil {
			return err
		}
	}

	daemonListener, err := net.Listen("unix", "/var/run/ravel.sock")
	if err != nil {
		return err
	}

	go d.server.Serve(daemonListener)

	return nil
}

func (d *Daemon) Run(runCtx context.Context) {
	<-runCtx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var waitGroup sync.WaitGroup

	waitGroup.Add(2)
	go func() {
		d.server.Shutdown(ctx)
		waitGroup.Done()
	}()

	go func() {
		if d.agent != nil {
			d.agent.Stop(ctx)
		}
		waitGroup.Done()
	}()

	waitGroup.Wait()
}
