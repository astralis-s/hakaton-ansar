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
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"

	"github.com/astralis-s/hakaton-ansar/api/openapi"
	"github.com/astralis-s/hakaton-ansar/internal/modules/catalog"
	"github.com/astralis-s/hakaton-ansar/internal/modules/crm"
	"github.com/astralis-s/hakaton-ansar/internal/modules/financing"
	financinginfra "github.com/astralis-s/hakaton-ansar/internal/modules/financing/infra"
	"github.com/astralis-s/hakaton-ansar/internal/modules/iam"
	"github.com/astralis-s/hakaton-ansar/internal/modules/scheduling"
	schedulingdomain "github.com/astralis-s/hakaton-ansar/internal/modules/scheduling/domain"
	schedulinginfra "github.com/astralis-s/hakaton-ansar/internal/modules/scheduling/infra"
	"github.com/astralis-s/hakaton-ansar/internal/platform/config"
	"github.com/astralis-s/hakaton-ansar/internal/platform/database"
	"github.com/astralis-s/hakaton-ansar/internal/platform/httpserver"
	"github.com/astralis-s/hakaton-ansar/internal/platform/logger"
	"github.com/astralis-s/hakaton-ansar/internal/platform/web"
	publicapiv1 "github.com/astralis-s/hakaton-ansar/internal/publicapi/v1"
	"github.com/astralis-s/hakaton-ansar/internal/seed"
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

	if cfg.SeedOnBoot {
		if err := seed.Run(ctx, pool, seed.Config{
			Lat:      cfg.Prayer.Lat,
			Lon:      cfg.Prayer.Lon,
			Madhab:   cfg.Prayer.Madhab,
			Method:   cfg.Prayer.Method,
			Timezone: loadTimezone(cfg.Prayer.Timezone, log),
		}, log); err != nil {
			return err
		}
	}

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
	catalogModule := catalog.New(catalog.Deps{Pool: pool, Log: log})
	crmModule := crm.New(crm.Deps{Pool: pool, Log: log})
	financingModule := financing.New(financing.Deps{
		Pool:                  pool,
		Tx:                    txManager,
		Log:                   log,
		ComparisonRatePercent: decimal.NewFromInt(int64(cfg.Financing.ComparisonRatePercent)),
		Products:              financinginfra.NewProductReader(catalogModule.Products()),
		Clients:               financinginfra.NewClientReader(crmModule.Clients()),
		OwnerOnly:             iamModule.OwnerMiddleware(),
	})

	prayerLoc := schedulingdomain.Location{Lat: cfg.Prayer.Lat, Lon: cfg.Prayer.Lon, TZ: loadTimezone(cfg.Prayer.Timezone, log)}
	schedulingModule := scheduling.New(scheduling.Deps{
		Pool:     pool,
		Log:      log,
		Provider: schedulinginfra.NewPrayerProvider(prayerLoc, cfg.Prayer.Madhab, cfg.Prayer.Method),
		Policy:   prayerPolicy(cfg.Prayer),
		Location: prayerLoc,
	})

	// Public API (+3) — thin layer over the same financing use-cases.
	publicAPI := publicapiv1.New(publicapiv1.Deps{
		CreateContract: financingModule.CreateContractUseCase(),
		GetContract:    financingModule.GetContractUseCase(),
		Log:            log,
	})

	srv := httpserver.New(httpserver.Config{
		Port:               cfg.HTTP.Port,
		ReadTimeout:        cfg.HTTP.ReadTimeout,
		WriteTimeout:       cfg.HTTP.WriteTimeout,
		ShutdownTimeout:    cfg.HTTP.ShutdownTimeout,
		CORSAllowedOrigins: cfg.HTTP.CORSAllowedOrigins,
	}, log)

	mountRoutes(srv.Router(), iamModule, catalogModule, crmModule, financingModule, schedulingModule, publicAPI)

	return srv.Run(ctx)
}

// loadTimezone loads an IANA timezone, falling back to a fixed MSK offset.
func loadTimezone(name string, log *slog.Logger) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		log.Warn("failed to load timezone, using MSK (UTC+3)", "timezone", name, "error", err)
		return time.FixedZone("MSK", 3*60*60)
	}
	return loc
}

// prayerPolicy builds the scheduling policy from config (Jummah HH:MM strings).
func prayerPolicy(p config.Prayer) schedulingdomain.Policy {
	parseTOD := func(s string, defH, defM int) schedulingdomain.TimeOfDay {
		t, err := time.Parse("15:04", s)
		if err != nil {
			return schedulingdomain.TimeOfDay{Hour: defH, Minute: defM}
		}
		return schedulingdomain.TimeOfDay{Hour: t.Hour(), Minute: t.Minute()}
	}
	return schedulingdomain.Policy{
		BufferBefore:  time.Duration(p.BufferBeforeMin) * time.Minute,
		BufferAfter:   time.Duration(p.BufferAfterMin) * time.Minute,
		JummahEnabled: true,
		JummahStart:   parseTOD(p.JummahStart, 12, 30),
		JummahEnd:     parseTOD(p.JummahEnd, 14, 0),
	}
}

// mountRoutes registers the health check, Swagger UI and the two API surfaces.
func mountRoutes(r chi.Router, iamModule *iam.Module, catalogModule *catalog.Module, crmModule *crm.Module, financingModule *financing.Module, schedulingModule *scheduling.Module, publicAPI *publicapiv1.Module) {
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
			catalogModule.RegisterRoutes(pr)
			crmModule.RegisterRoutes(pr)
			financingModule.RegisterRoutes(pr)
			schedulingModule.RegisterRoutes(pr)
		})
	})

	// Public API (+3) — X-API-Key.
	r.Route("/api/v1", func(vr chi.Router) {
		vr.Use(iamModule.APIKeyMiddleware())
		publicAPI.RegisterRoutes(vr)
	})
}
