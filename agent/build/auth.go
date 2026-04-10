package build

import (
	"context"
	"sync"

	"github.com/alexisbouchez/ravel/core/registry"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// authProvider implements BuildKit's session.Attachable interface for registry auth
type authProvider struct {
	registries registry.RegistriesConfig
	mu         sync.Mutex
}

// newAuthProvider creates a new auth provider from Ravel's registry config
func newAuthProvider(registries registry.RegistriesConfig) session.Attachable {
	return &authProvider{
		registries: registries,
	}
}

// Register implements session.Attachable
func (ap *authProvider) Register(server *grpc.Server) {
	auth.RegisterAuthServer(server, ap)
}

// Credentials implements auth.AuthServer
func (ap *authProvider) Credentials(ctx context.Context, req *auth.CredentialsRequest) (*auth.CredentialsResponse, error) {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	host := req.Host
	if ap.registries == nil {
		return &auth.CredentialsResponse{}, nil
	}

	username, password, err := ap.registries.Get(host)
	if err != nil {
		return nil, err
	}

	return &auth.CredentialsResponse{
		Username: username,
		Secret:   password,
	}, nil
}

// FetchToken implements auth.AuthServer - not supported, return unimplemented
func (ap *authProvider) FetchToken(ctx context.Context, req *auth.FetchTokenRequest) (*auth.FetchTokenResponse, error) {
	return nil, status.Error(codes.Unimplemented, "FetchToken not supported")
}

// GetTokenAuthority implements auth.AuthServer - not supported, return unimplemented
func (ap *authProvider) GetTokenAuthority(ctx context.Context, req *auth.GetTokenAuthorityRequest) (*auth.GetTokenAuthorityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetTokenAuthority not supported")
}

// VerifyTokenAuthority implements auth.AuthServer - not supported, return unimplemented
func (ap *authProvider) VerifyTokenAuthority(ctx context.Context, req *auth.VerifyTokenAuthorityRequest) (*auth.VerifyTokenAuthorityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "VerifyTokenAuthority not supported")
}
