package edge

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/valyentdev/corroclient"
)

type Backends struct {
	corro    *corroclient.CorroClient
	backends map[string]*Backend
	mutex    sync.RWMutex
}

func newBackends(corro *corroclient.CorroClient) *Backends {
	return &Backends{
		corro:    corro,
		backends: make(map[string]*Backend),
	}
}

func (b *Backends) Start() {
	go func() {
		for {
			err := b.sync()
			if err != nil {
				slog.Error("error during syncing", "err", err)
			}

			time.Sleep(2 * time.Second)
		}
	}()
}

func scanInstance(row *corroclient.Row) (Instance, error) {
	var i Instance
	err := row.Scan(&i.Id, &i.Gatewayid, &i.TargetPort, &i.Address, &i.Port)
	if err != nil {
		return Instance{}, err
	}

	return i, nil
}

const getBackendsQuery = `
						  select i.id as id, g.id, g.target_port, n.address, n.http_proxy_port  
						  from instances i
						  join machines m on m.instance_id = i.id
						  join gateways g on g.fleet_id = m.fleet_id 
						  join nodes n on n.id = i.node
						  where i.status = 'running'
						`

func (b *Backends) sync() error {
	sub, err := b.corro.PostSubscription(context.Background(), corroclient.Statement{
		Query: getBackendsQuery,
	})
	if err != nil {
		return err
	}

	events := sub.Events()
	for e := range events {
		switch e.Type() {
		case corroclient.EventTypeRow:
			row := e.(*corroclient.Row)
			i, err := scanInstance(row)
			if err != nil {
				slog.Error("error scanning instance", "err", err)
				continue
			}

			b.addBackend(i)
		case corroclient.EventTypeChange:
			change := e.(*corroclient.Change)
			i, err := scanInstance(change.Row)
			if err != nil {
				slog.Error("error scanning instance", "err", err)
				continue
			}

			switch change.ChangeType {
			case corroclient.ChangeTypeInsert:
				b.addBackend(i)
			case corroclient.ChangeTypeDelete:
				b.removeBackend(i.Id)
			}
		}
	}

	return nil
}

func (b *Backends) addBackend(i Instance) {
	b.mutex.Lock()
	b.backends[i.Id] = newBackend(i)
	b.mutex.Unlock()
}

func (b *Backends) removeBackend(id string) {
	b.mutex.Lock()
	delete(b.backends, id)
	b.mutex.Unlock()
}

func (b *Backends) getBackend(id string) (*Backend, bool) {
	b.mutex.RLock()
	backend, ok := b.backends[id]
	b.mutex.RUnlock()

	if !ok {
		return nil, false
	}

	return backend, true
}
