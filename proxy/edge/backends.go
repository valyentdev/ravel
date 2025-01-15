package edge

import (
	"context"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/valyentdev/corroclient"
)

type instance struct {
	Id                   string `json:"id"`
	Namespace            string `json:"namespace"`
	MachineId            string `json:"machine_id"`
	Address              string `json:"address"`
	Port                 int    `json:"port"`
	EnableMachineGateway bool   `json:"enable_machine_gateway"`
}

func (i *instance) Url() *url.URL {
	url := &url.URL{
		Scheme: "http",
		Host:   strings.Join([]string{i.Address, strconv.Itoa(i.Port)}, ":"),
	}

	slog.Debug("instance url", "url", url)
	return url
}

type instanceBackends struct {
	corro    *corroclient.CorroClient
	backends map[string]*instance
	mutex    sync.RWMutex
}

func newInstanceBackends(corro *corroclient.CorroClient) *instanceBackends {
	return &instanceBackends{
		corro:    corro,
		backends: make(map[string]*instance),
	}
}

func (b *instanceBackends) Start() {
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

func scanInstance(row *corroclient.Row) (instance, error) {
	var i instance
	var machineGatewayEnabled uint8

	err := row.Scan(&i.Id, &i.MachineId, &i.Namespace, &i.Address, &i.Port, &machineGatewayEnabled)
	if err != nil {
		return instance{}, err
	}

	i.EnableMachineGateway = machineGatewayEnabled == 1

	return i, nil
}

const getBackendsQuery = `
						  select i.id, m.id, m.namespace, n.address, n.http_proxy_port, i.enable_machine_gateway
						  from instances i
						  join machines m on m.instance_id = i.id
						  join nodes n on n.id = i.node
						  where i.status = 'running'
						`

func (b *instanceBackends) sync() error {
	sub, err := b.corro.Subscribe(context.Background(), corroclient.Statement{
		Query: getBackendsQuery,
	})
	if err != nil {
		return err
	}

	for {
		e, err := sub.Next()
		if err != nil {
			return err
		}
		switch e.Type() {
		case corroclient.EventTypeRow:
			row := e.(*corroclient.Row)
			i, err := scanInstance(row)
			if err != nil {
				slog.Error("error scanning instance", "err", err)
				continue
			}

			b.setBackend(i)
		case corroclient.EventTypeChange:
			change := e.(*corroclient.Change)
			i, err := scanInstance(change.Row)
			if err != nil {
				slog.Error("error scanning instance", "err", err)
				continue
			}

			switch change.ChangeType {
			case corroclient.ChangeTypeUpdate:
				b.setBackend(i)
			case corroclient.ChangeTypeInsert:
				b.setBackend(i)
			case corroclient.ChangeTypeDelete:
				b.deleteBackend(i.MachineId)
			}
		}
	}
}

func (b *instanceBackends) setBackend(i instance) {
	slog.Debug("setting backend", "instance", i)
	b.mutex.Lock()
	b.backends[i.MachineId] = &i
	b.mutex.Unlock()
}

func (b *instanceBackends) deleteBackend(id string) {
	slog.Debug("deleting backend", "id", id)
	b.mutex.Lock()
	delete(b.backends, id)
	b.mutex.Unlock()
}

func (b *instanceBackends) getBackend(mid string) (*instance, bool) {
	b.mutex.RLock()
	backend, ok := b.backends[mid]
	b.mutex.RUnlock()
	if !ok {
		return nil, false
	}
	return backend, true
}
