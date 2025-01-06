package api

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/valyentdev/ravel/initd"
	"github.com/valyentdev/ravel/initd/environment"
	"github.com/valyentdev/ravel/initd/files"
	"github.com/valyentdev/ravel/pkg/vsock"
)

func ServeInitdAPI(env *environment.Env) error {
	publicEndpoints := &publicEndpoints{
		files: &files.Service{},
	}

	publicMux := http.NewServeMux()
	publicAPI := humago.New(publicMux, huma.DefaultConfig("Initd API", "1.0.0"))
	publicEndpoints.registerRoutes(publicAPI)

	internalEndpoints := &InternalEndpoint{
		env: env,
	}

	internalMux := http.NewServeMux()
	internalAPI := humago.New(internalMux, huma.DefaultConfig("Initd Internal API", "1.0.0"))
	publicEndpoints.registerRoutes(internalAPI)
	internalEndpoints.registerRoutes(internalAPI)

	vsockLn, err := vsock.Listener(context.Background(), nil, 10000)
	if err != nil {
		return fmt.Errorf("failed to create vsock listener: %w", err)
	}

	publicLn, err := net.Listen("tcp", fmt.Sprintf(":%d", initd.InitdPort))
	if err != nil {
		return fmt.Errorf("failed to create public listener: %w", err)
	}

	go http.Serve(publicLn, publicMux)
	http.Serve(vsockLn, internalMux)

	return nil
}
