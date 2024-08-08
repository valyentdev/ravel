package server

import (
	"github.com/danielgtaylor/huma/v2/adapters/humafiber"
	"github.com/gofiber/fiber/v2"
	"github.com/valyentdev/ravel/internal/server/endpoints/machines"
	"github.com/valyentdev/ravel/internal/server/endpoints/nodes"
	"github.com/valyentdev/ravel/internal/server/utils"
)

func (s *Server) registerEndpoints(f *fiber.App) {
	humaConfig := utils.GetHumaConfig()

	api := humafiber.New(f, humaConfig)

	machines.NewEndpoint(s.m, s.validator).Register(api)
	nodes.NewEndpoint(s.m).Register(api)
}
