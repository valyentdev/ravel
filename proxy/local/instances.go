package local

import (
	"context"
	"log/slog"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/valyentdev/corroclient"
)

type instances struct {
	nodeId    string
	instances map[string]*Instance
	mutex     sync.RWMutex
	corro     *corroclient.CorroClient
}

func newInstances(corro *corroclient.CorroClient, nodeId string) *instances {
	return &instances{
		nodeId:    nodeId,
		instances: make(map[string]*Instance),
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
	InstanceId string
	Ip         string
}

func (i *Instance) Url(port string) *url.URL {
	return &url.URL{
		Scheme: "http",
		Host:   strings.Join([]string{i.Ip, port}, ":"),
	}
}

func scanInstance(row *corroclient.Row) (Instance, error) {
	var i Instance
	err := row.Scan(&i.InstanceId, &i.Ip)
	if err != nil {
		return i, err
	}

	return i, nil
}

const getInstancesQuery = `
						  select i.id, i.local_ipv4
						  from instances i
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
		case corroclient.EventTypeChange:
			change := e.(*corroclient.Change)
			ie, err := scanInstance(change.Row)
			if err != nil {
				slog.Error("error scanning instance", "err", err)
				continue
			}
			switch change.ChangeType {
			case corroclient.ChangeTypeInsert, corroclient.ChangeTypeUpdate:
				i.addInstance(ie)
			case corroclient.ChangeTypeDelete:
				i.removeInstance(ie.InstanceId)
			}

		}
	}
	return nil
}

func (i *instances) addInstance(ie Instance) {
	i.mutex.Lock()
	i.instances[ie.InstanceId] = &ie
	i.mutex.Unlock()
}

func (i *instances) removeInstance(instanceId string) {
	i.mutex.Lock()
	delete(i.instances, instanceId)
	i.mutex.Unlock()
}

func (i *instances) getInstance(instanceId string) (*Instance, bool) {
	i.mutex.RLock()
	b, ok := i.instances[instanceId]
	i.mutex.RUnlock()
	if !ok {
		return nil, false
	}

	return b, ok
}
