package edge

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/valyentdev/corroclient"
)

type Gateways struct {
	gateways        map[string]Gateway
	gatewaysDomains map[string]string
	mutex           sync.RWMutex
	corro           *corroclient.CorroClient
}

func newGateways(corro *corroclient.CorroClient) *Gateways {
	return &Gateways{
		gateways:        make(map[string]Gateway),
		gatewaysDomains: make(map[string]string),
		corro:           corro,
	}
}

func (g *Gateways) Start() {
	go func() {
		for {
			err := g.sync()
			slog.Warn("syncing stopped")
			if err != nil {
				slog.Error("error during syncing", "err", err)
			}

			time.Sleep(2 * time.Second)
		}
	}()
}

const getConfigQuery = `select g.id, g.namespace, g.name, g.target_port, 
						json_group_array(i.id) as targets  
						from machines m  
						join gateways g on g.fleet_id = m.fleet_id 
					   	join instances i on i.machine_id = m.id
						where i.status = 'running'
						group by g.id
					`

func scanGateway(row *corroclient.Row) (Gateway, error) {
	var gw Gateway
	var instanceBytes []byte
	err := row.Scan(&gw.Id, &gw.Namespace, &gw.Name, &gw.TargetPort, &instanceBytes)
	if err != nil {
		return gw, err
	}

	err = json.Unmarshal(instanceBytes, &gw.Instances)
	if err != nil {
		return gw, err
	}

	return gw, err
}

func (g *Gateways) sync() error {
	sub, err := g.corro.PostSubscription(context.Background(), corroclient.Statement{
		Query: getConfigQuery,
	})
	if err != nil {
		return err
	}

	events := sub.Events()
	for e := range events {
		switch e.Type() {
		case corroclient.EventTypeRow:
			row := e.(*corroclient.Row)
			gw, err := scanGateway(row)
			if err != nil {
				slog.Error("error during gateway scanning", "err", err)
				continue
			}

			g.addGateway(gw)
		case corroclient.EventTypeChange:
			change := e.(*corroclient.Change)
			gw, err := scanGateway(change.Row)
			if err != nil {
				slog.Error("error during gateway scanning", "err", err)
				continue
			}

			switch change.ChangeType {
			case corroclient.ChangeTypeInsert:
				g.addGateway(gw)
			case corroclient.ChangeTypeUpdate:
				g.updateGateway(gw)
			case corroclient.ChangeTypeDelete:
				g.removeGateway(gw)
			}
		}
	}

	return nil
}

func getDomain(name, namespace string) string {
	return name + "-" + namespace
}

func (g *Gateways) addGateway(gw Gateway) {
	g.mutex.Lock()
	g.gateways[gw.Id] = gw
	g.gatewaysDomains[getDomain(gw.Name, gw.Namespace)] = gw.Id
	g.mutex.Unlock()
}

func (g *Gateways) removeGateway(gw Gateway) {
	g.mutex.Lock()
	delete(g.gateways, gw.Id)
	delete(g.gatewaysDomains, getDomain(gw.Name, gw.Namespace))
	g.mutex.Unlock()
}

func (g *Gateways) updateGateway(gw Gateway) {
	g.mutex.Lock()
	old := g.gateways[gw.Id]
	delete(g.gatewaysDomains, getDomain(old.Name, old.Namespace))
	g.gateways[gw.Id] = gw
	g.gatewaysDomains[getDomain(gw.Name, gw.Namespace)] = gw.Id
	g.mutex.Unlock()
}

func (g *Gateways) GetGateway(domain string) (*Gateway, bool) {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	id, ok := g.gatewaysDomains[domain]
	if !ok {
		return nil, false
	}
	gw, ok := g.gateways[id]
	if !ok {
		return nil, false
	}

	return &gw, true
}
