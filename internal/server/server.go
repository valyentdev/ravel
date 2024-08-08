package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/valyentdev/ravel/internal/server/utils"
	"github.com/valyentdev/ravel/pkg/config"
	"github.com/valyentdev/ravel/pkg/manager"
)

type Server struct {
	m         *manager.Manager
	fiber     *fiber.App
	validator *utils.Validator
}

func NewServer(c config.RavelConfig) (*Server, error) {
	app := fiber.New()
	m, err := manager.New(manager.ManagerConfig{
		CorrosionConfig: c.Corrosion,
		NatsURl:         c.Nats.Url,
	})
	if err != nil {
		return nil, err
	}

	return &Server{
		fiber:     app,
		validator: utils.NewValidator(),
		m:         m,
	}, nil
}

func (s *Server) Serve() error {
	s.registerEndpoints(s.fiber)
	return s.fiber.Listen(":3000")
}

func (s *Server) Stop() error {
	return s.fiber.Shutdown()
}
