package hcfiber

import (
	"context"

	"github.com/gofiber/fiber/v2"

	svchealthcheck "github.com/jamillosantos/services-healthcheck"
)

// Healthchecker abstracts the implementation of the svchealthcheck.Healthcheck.
type Healthchecker interface {
	Health(ctx context.Context) *svchealthcheck.CheckResponse
	Ready(ctx context.Context) *svchealthcheck.CheckResponse
}

// FiberApp abstracts the implementation of the fiber.App.
type FiberApp interface {
	Get(path string, handlers ...fiber.Handler) fiber.Router
}

// FiberInitialize that will set up the endpoints on a given fiber.App.
func FiberInitialize(healthcheck Healthchecker, app FiberApp) {
	app.Get(svchealthcheck.HealthPath, fiberEndpoint(healthcheck.Health))
	app.Get(svchealthcheck.ReadyPath, fiberEndpoint(healthcheck.Ready))
}

func fiberEndpoint(getResponse func(ctx context.Context) *svchealthcheck.CheckResponse) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		r := getResponse(ctx.Context())
		return ctx.Status(r.StatusCode).JSON(r)
	}
}
