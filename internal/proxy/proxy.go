package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/valyentdev/corroclient"
	"github.com/valyentdev/ravel/pkg/core"
)

type RavelProxy struct {
	gateways map[string]Gateway
	mutex    sync.RWMutex
	corro    *corroclient.CorroClient
}

func NewRavelProxy(corro *corroclient.CorroClient) *RavelProxy {
	return &RavelProxy{
		corro: corro,
	}
}

const getConfigQuery = `select g.id, g.namespace, g.name, g.target_port, json_group_array(i.local_ipv4) as targets  
					   from machines m  
					   join gateways g on g.fleet_id = m.fleet_id 
					   join instances i on i.machine_id = m.id 
					   where i.status = 'running'
					   group by g.id
					   `

type Gateway struct {
	core.Gateway
	Instances []string
}

func (p *RavelProxy) Start() {
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		for range ticker.C {

			err := p.refreshConfig()
			if err != nil {
				fmt.Println(err)
			}

		}
	}()

}

func (p *RavelProxy) refreshConfig() error {
	rows, err := p.corro.Query(context.Background(), corroclient.Statement{
		Query: getConfigQuery,
	})
	if err != nil {
		return err
	}

	var clusters []Gateway

	for rows.Next() {
		var g Gateway
		var instancesJSON []byte

		err := rows.Scan(&g.Id, &g.Namespace, &g.Name, &g.TargetPort, &instancesJSON)
		if err != nil {
			return err
		}

		err = json.Unmarshal(instancesJSON, &g.Instances)
		if err != nil {
			return err
		}

		clusters = append(clusters, g)
	}

	bytes, err := json.Marshal(clusters)
	if err != nil {
		return err
	}

	newConfig := make(map[string]Gateway)

	for _, cluster := range clusters {
		newConfig[cluster.Name+"-"+cluster.Namespace] = cluster
	}

	p.mutex.Lock()
	p.gateways = newConfig
	p.mutex.Unlock()

	fmt.Println(string(bytes))
	return nil
}

func (p *RavelProxy) Handle(w http.ResponseWriter, r *http.Request) {
	host := r.Host

	if strings.HasSuffix(host, ".valyent.app") {
		host = strings.TrimSuffix(host, ".valyent.app")
	} else {
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	p.mutex.RLock()
	gateway, ok := p.gateways[host]
	p.mutex.RUnlock()

	if !ok {
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	if len(gateway.Instances) == 0 {
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	// pick a random instance
	instance := gateway.Instances[0]

	// proxy the request

	httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", instance, gateway.TargetPort),
	}).ServeHTTP(w, r)

}
