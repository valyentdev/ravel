package api

import (
	"context"
	"net/http"

	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/core/api/httpclient"
)

type RavelClient struct {
	client *httpclient.Client
	config RavelClientConfig
}

type opt struct {
	namespace     string
	authorization string
}

type Opt func(*opt)

func WithNamespace(namespace string) Opt {
	return func(o *opt) {
		o.namespace = namespace
	}
}

func WithAuthorization(authorization string) Opt {
	return func(o *opt) {
		o.authorization = authorization
	}
}

type RavelClientConfig struct {
	ApiUrl           string // Default: https://api.valyent.cloud
	Authorization    string // Default: ""
	DefaultNamespace string // Default: "" (to specify only on custom ravel cluster and not on valyent api)
}

func NewRavelClient(config RavelClientConfig) *RavelClient {

	return &RavelClient{
		client: httpclient.NewClient(config.ApiUrl, http.DefaultClient),
		config: config,
	}
}

func (rc *RavelClient) getReqOpts(opts []Opt, includeNS bool) []httpclient.ReqOpt {
	opt := &opt{
		namespace:     rc.config.DefaultNamespace,
		authorization: rc.config.Authorization,
	}

	for _, o := range opts {
		o(opt)
	}

	reqOpts := []httpclient.ReqOpt{}
	if includeNS && opt.namespace != "" {
		reqOpts = append(reqOpts, httpclient.WithQuery("namespace", opt.namespace))
	}

	if opt.authorization != "" {
		reqOpts = append(reqOpts, httpclient.WithHeader("Authorization", opt.authorization))
	}

	return reqOpts

}

func (rc *RavelClient) ListNamespaces(ctx context.Context, opts ...Opt) ([]Namespace, error) {
	var namespaces []Namespace
	err := rc.client.Get(ctx, "/namespaces", &namespaces, rc.getReqOpts(opts, false)...)
	if err != nil {
		return nil, err
	}

	return namespaces, nil
}

type CreateNamespaceBody struct {
	Name string `json:"name"`
}

func (rc *RavelClient) CreateNamespace(ctx context.Context, name string, opts ...Opt) (*Namespace, error) {
	var namespace Namespace
	body := CreateNamespaceBody{
		Name: name,
	}
	err := rc.client.Post(ctx, "/namespaces", body, &namespace, rc.getReqOpts(opts, false)...)
	if err != nil {
		return nil, err
	}

	return &namespace, nil
}

func (rc *RavelClient) GetNamespace(ctx context.Context, namespace string, opts ...Opt) (*Namespace, error) {
	var ns Namespace
	err := rc.client.Get(ctx, "/namespaces/"+namespace, &ns, rc.getReqOpts(opts, false)...)
	if err != nil {
		return nil, err
	}

	return &ns, nil
}

type CreateFleetBody struct {
	Name string `json:"name"`
}

func (rc *RavelClient) CreateFleet(ctx context.Context, name string, opts ...Opt) (*core.Fleet, error) {
	opt := &opt{}
	for _, o := range opts {
		o(opt)
	}

	var fleet core.Fleet

	body := CreateFleetBody{
		Name: name,
	}

	err := rc.client.Post(ctx, "/fleets", body, &fleet)
	if err != nil {
		return nil, err
	}

	return &fleet, nil
}

func (rc *RavelClient) ListFleets(ctx context.Context, namespace string) ([]core.Fleet, error) {
	var fleets []core.Fleet
	err := rc.client.Get(ctx, "/namespaces/"+namespace+"/fleets", &fleets)
	if err != nil {
		return nil, err
	}

	return fleets, nil
}

func (rc *RavelClient) ListMachines(ctx context.Context, namespace string) ([]Machine, error) {
	var machines []Machine
	err := rc.client.Get(ctx, "/fleets/"+namespace+"/machines", &machines)
	if err != nil {
		return nil, err
	}

	return machines, nil
}
