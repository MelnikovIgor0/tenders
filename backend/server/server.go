package server

import (
	"backend/config"
	"backend/handlers"
	"backend/storage"
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"go.uber.org/fx"
)

func buildFiberServer(lc fx.Lifecycle, h *handlers.Handlers, c *config.Config) *fiber.App {
	app := fiber.New()
	app.Use(cors.New())
	app.Use(logger.New())

	api := app.Group("/api")
	api.Get("/ping", h.Ping)
	tenders := api.Group("/tenders")
	tenders.Post("/new", h.CreateTender)
	tenders.Get("/", h.FilterTenders)
	tenders.Get("/my", h.FilterMyTenders)
	tendersCRUD := tenders.Group("/:tenderId")
	tendersCRUD.Get("/status", h.GetTenderStatus)
	tendersCRUD.Put("/status", h.UpdateTenderStatus)
	tendersCRUD.Patch("/edit", h.EditTender)
	bids := api.Group("/bids")
	bids.Post("/new", h.CreateBid)
	bids.Get("/my", h.GetMyBids)
	tendersCRUD.Get("/list", h.GetTenderBids)
	bidsCRUD := bids.Group("/:bidId")
	bidsCRUD.Get("/status", h.GetBidStatus)
	bidsCRUD.Put("/status", h.ChangeBidStatus)
	bidsCRUD.Patch("/edit", h.EditBid)
	bidsCRUD.Put("/submit_decision", h.SetDecision)
	bidsCRUD.Get("/get_decision", h.GetDecision)

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go app.Listen(c.GetServerAddress())
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return app.Shutdown()
		},
	})

	return app
}

func BuildServerAndEnv() *fx.App {
	return fx.New(
		fx.Provide(
			config.NewConfig,
			storage.NewStorage,
			handlers.NewHandlers,
		),
		fx.Invoke(buildFiberServer),
	)
}
