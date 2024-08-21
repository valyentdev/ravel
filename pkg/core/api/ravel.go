package api

import (
	"context"
	"net/http"

	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/core/api/httpclient"
)

type RavelClient struct {
	client *httpclient.Client
}

type ravelClientOptions struct {
	apiUrl           string
	defaultNamespace string
}

func WithDefaultNamespace(namespace string) RavelClientOpt {
	return func(o *ravelClientOptions) {
		o.defaultNamespace = namespace
	}
}

type RavelClientOpt func(*ravelClientOptions)

func WithApiUrl(url string) RavelClientOpt {
	return func(o *ravelClientOptions) {
		o.apiUrl = url
	}
}

type opt struct {
	namespace string
}

type Opt func(*opt)

func WithNamespace(namespace string) Opt {
	return func(o *opt) {
		o.namespace = namespace
	}
}

func NewRavelClient(opts ...RavelClientOpt) *RavelClient {
	opt := &ravelClientOptions{
		apiUrl: "https://api.ravel.sh",
	}

	for _, o := range opts {
		o(opt)
	}

	return &RavelClient{client: httpclient.NewClient(opt.apiUrl, http.DefaultClient)}
}

func (rc *RavelClient) ListNamespaces(ctx context.Context) ([]Namespace, error) {
	var namespaces []Namespace
	err := rc.client.Get(ctx, "/namespaces", &namespaces)
	if err != nil {
		return nil, err
	}

	return namespaces, nil
}

type CreateNamespaceBody struct {
	Name string `json:"name"`
}

func (rc *RavelClient) CreateNamespace(ctx context.Context, name string) (*Namespace, error) {
	var namespace Namespace
	body := CreateNamespaceBody{
		Name: name,
	}
	err := rc.client.Post(ctx, "/namespaces", body, &namespace)
	if err != nil {
		return nil, err
	}

	return &namespace, nil
}

func (rc *RavelClient) GetNamespace(ctx context.Context, namespace string) (*Namespace, error) {
	var ns Namespace
	err := rc.client.Get(ctx, "/namespaces/"+namespace, &ns)
	if err != nil {
		return nil, err
	}

	return &ns, nil
}

func (rc *RavelClient) getReqOpt(opts []Opt) []httpclient.ReqOpt {
	opt := opt{}
	for _, o := range opts {
		o(&opt)
	}

	reqOpts := []httpclient.ReqOpt{}

	if opt.namespace != "" {
		reqOpts = append(reqOpts, httpclient.WithQuery("namespace", opt.namespace))
	}

	return reqOpts
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
