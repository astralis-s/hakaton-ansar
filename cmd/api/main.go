// Command api is the single entry point of the Amana modular monolith. It loads
// configuration, runs migrations, opens the database pool, wires the modules
// onto the two HTTP surfaces (/api/app, /api/v1) and starts the server. All
// dependency wiring lives here and nowhere else (ARCHITECTURE.md: DI).
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"

	"github.com/astralis-s/hakaton-ansar/api/openapi"
	"github.com/astralis-s/hakaton-ansar/internal/modules/iam"
	"github.com/astralis-s/hakaton-ansar/internal/platform/config"
	"github.com/astralis-s/hakaton-ansar/internal/platform/database"
	"github.com/astralis-s/hakaton-ansar/internal/platform/httpserver"
	"github.com/astralis-s/hakaton-ansar/internal/platform/logger"
	"github.com/astralis-s/hakaton-ansar/internal/platform/web"
	"github.com/astralis-s/hakaton-ansar/migrations"
)

// @title           Amana Public API
// @version         0.1.0
// @description     Публичный API CRM «Амана» (мурабаха-рассрочка без рибы).
// @BasePath        /api/v1
// @securityDefinitions.apikey  ApiKeyAuth
// @in              header
// @name            X-API-Key
func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	migrateOnly := flag.Bool("migrate-only", false, "apply database migrations and exit")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log := logger.New(cfg.Logger.Level, cfg.Logger.Format)
	slog.SetDefault(log)
	log.Info("starting amana backend", "env", cfg.Env, "port", cfg.HTTP.Port)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if cfg.MigrateOnBoot || *migrateOnly {
		log.Info("applying database migrations")
		if err := database.Migrate(ctx, cfg.DB.URL, migrations.FS, log); err != nil {
			return fmt.Errorf("run migrations: %w", err)
		}
		if *migrateOnly {
			log.Info("migrations applied; exiting (migrate-only)")
			return nil
		}
	}

	pool, err := database.NewPool(ctx, cfg.DB)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer pool.Close()
	log.Info("database connected")

	// Shared infrastructure.
	txManager := database.NewTxManager(pool)

	// Modules.
	iamModule := iam.New(iam.Deps{
		Pool:      pool,
		Tx:        txManager,
		Log:       log,
		JWTSecret: cfg.Auth.JWTSecret,
		JWTTTL:    cfg.Auth.JWTTTL,
	})

	srv := httpserver.New(httpserver.Config{
		Port:               cfg.HTTP.Port,
		ReadTimeout:        cfg.HTTP.ReadTimeout,
		WriteTimeout:       cfg.HTTP.WriteTimeout,
		ShutdownTimeout:    cfg.HTTP.ShutdownTimeout,
		CORSAllowedOrigins: cfg.HTTP.CORSAllowedOrigins,
	}, log)

	mountRoutes(srv.Router(), iamModule)

	return srv.Run(ctx)
}

// mountRoutes registers the health check, Swagger UI and the two API surfaces.
func mountRoutes(r chi.Router, iamModule *iam.Module) {
	// Liveness probe — always 200 once the server is up.
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		web.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Swagger UI for the public API.
	httpserver.MountSwagger(r, openapi.SpecJSON)

	// Internal application API (SPA) — JWT.
	r.Route("/api/app", func(ar chi.Router) {
		// Public: login + first-run setup.
		iamModule.RegisterPublicAppRoutes(ar)

		// Protected: everything else lives behind JWT auth.
		ar.Group(func(pr chi.Router) {
			pr.Use(iamModule.JWTMiddleware())
			iamModule.RegisterProtectedAppRoutes(pr)
			// Phase 3+: catalog, crm, financing, scheduling mount here.
		})
	})

	// Public API (+3) — X-API-Key.
	r.Route("/api/v1", func(vr chi.Router) {
		vr.Use(iamModule.APIKeyMiddleware())
		vr.Get("/ping", func(w http.ResponseWriter, _ *http.Request) {
			web.JSON(w, http.StatusOK, map[string]string{"surface": "v1", "status": "ok"})
		})
		// Phase 6: publicapi mounts here.
	})
}
