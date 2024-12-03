package local

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/valyentdev/corroclient"
)

type instances struct {
	nodeId    string
	instances map[string]*Backend
	mutex     sync.RWMutex
	corro     *corroclient.CorroClient
}

func newInstances(corro *corroclient.CorroClient, nodeId string) *instances {
	return &instances{
		nodeId:    nodeId,
		instances: make(map[string]*Backend),
		corro:     corro,
	}
}

func (i *instances) Start() {
	go func() {
		for {
			err := i.sync()
			if err != nil {
				slog.Error("error during syncing", "err", err)
			}

			time.Sleep(2 * time.Second)
		}
	}()
}

type Instance struct {
	GatewayId  string
	InstanceId string
	Ip         string
	Port       int
}

func scanInstance(row *corroclient.Row) (Instance, error) {
	var i Instance
	err := row.Scan(&i.InstanceId, &i.GatewayId, &i.Ip, &i.Port)
	if err != nil {
		return i, err
	}

	return i, nil
}

const getInstancesQuery = `
						  select i.id, gw.id, i.local_ipv4, gw.target_port
						  from instances i
						  join machines m on m.id = i.machine_id
						  join gateways gw on gw.fleet_id = m.fleet_id
						  where i.status = 'running' AND i.node = ?
						`

func (i *instances) sync() error {
	sub, err := i.corro.PostSubscription(context.Background(), corroclient.Statement{
		Query:  getInstancesQuery,
		Params: []interface{}{i.nodeId},
	})
	if err != nil {
		return err
	}

	events := sub.Events()
	for e := range events {
		switch e.Type() {
		case corroclient.EventTypeRow:
			row := e.(*corroclient.Row)
			ie, err := scanInstance(row)
			if err != nil {
				slog.Error("error scanning instance", "err", err)
				continue
			}

			i.addInstance(ie)
		}
	}

	return nil
}

func (i *instances) addInstance(ie Instance) {
	i.mutex.Lock()
	i.instances[ie.GatewayId] = newBackend(ie)
	i.mutex.Unlock()
}

func (i *instances) getInstance(gw string) (*Backend, bool) {
	i.mutex.RLock()
	b, ok := i.instances[gw]
	i.mutex.RUnlock()
	if !ok {
		return nil, false
	}

	return b, ok
}
