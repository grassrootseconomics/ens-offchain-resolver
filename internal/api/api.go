package api

import (
	"context"
	"crypto"
	"log/slog"
	"net/http"
	"os"

	"github.com/grassrootseconomics/ens-offchain-resolver/pkg/ens"
	"github.com/kamikazechaser/common/httputil"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/bunrouter/extra/reqlog"
)

type (
	APIOpts struct {
		VerifyingKey  crypto.PublicKey
		EnableMetrics bool
		ListenAddress string
		Logg          *slog.Logger
		ENSProvider   *ens.ENS
	}

	API struct {
		validator    httputil.ValidatorProvider
		verifyingKey crypto.PublicKey
		router       *bunrouter.Router
		server       *http.Server
		logg         *slog.Logger
		ensProvider  *ens.ENS
	}
)

const apiVersion = "/api/v1"

func New(o APIOpts) *API {
	api := &API{
		validator:    httputil.NewValidator(""),
		verifyingKey: o.VerifyingKey,
		logg:         o.Logg,
		router: bunrouter.New(
			bunrouter.WithNotFoundHandler(notFoundHandler),
			bunrouter.WithMethodNotAllowedHandler(methodNotAllowedHandler),
		),
		ensProvider: o.ENSProvider,
	}

	if o.EnableMetrics {
		api.router.GET("/metrics", metricsHandler)
	}

	api.router.WithGroup(apiVersion, func(g *bunrouter.Group) {
		if os.Getenv("DEV") != "" {
			g = g.Use(reqlog.NewMiddleware())
		}

		g.GET("/:sender/*data", api.ccipHandler)
		g.GET("/resolve/:name", api.resolveNameHandler)
		g.GET("/reverse/:address", api.reverseAddressHandler)

		g.WithGroup("/register", func(rG *bunrouter.Group) {
			rG = rG.Use(api.authMiddleware)
			rG.POST("/", api.registerHandler)
		})
	})

	api.server = &http.Server{
		Addr:    o.ListenAddress,
		Handler: api.router,
	}

	return api
}

func (a *API) Start() error {
	a.logg.Info("API server starting", "address", a.server.Addr)
	if err := a.server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (a *API) Stop(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}
