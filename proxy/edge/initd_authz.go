package edge

import (
	"net/http"

	"github.com/valyentdev/ravel/internal/httpclient"
)

type Authorizer interface {
	Authorize(r *http.Request, namespace string) bool
}

type valyentAuthorizer struct {
	client   *httpclient.Client
	endpoint string
}

type authResult struct {
	Authorized bool   `json:"authorized"`
	Namespace  string `json:"namespace"`
}

func newValyentAuthorizer(endpoint string) *valyentAuthorizer {
	client := httpclient.NewClient(endpoint, http.DefaultClient)
	return &valyentAuthorizer{
		client:   client,
		endpoint: endpoint,
	}
}

func (v *valyentAuthorizer) Authorize(r *http.Request, namespace string) bool {
	var result authResult
	err := v.client.Get(
		r.Context(),
		v.endpoint,
		&result,
		httpclient.WithHeader("Authorization", r.Header.Get("Authorization")),
	)
	if err != nil {
		return false
	}

	if !result.Authorized {
		return false
	}

	return result.Namespace == namespace
}

type noAuthAuthorizer struct{}

func (n *noAuthAuthorizer) Authorize(r *http.Request, namespace string) bool {
	return true
}
