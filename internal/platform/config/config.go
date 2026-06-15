// Package config loads all runtime configuration from the environment.
// Secrets never live in code — only in env (CLAUDE.md §3).
package config

import (
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Config is the root configuration tree, grouped by concern.
type Config struct {
	Env           string `env:"APP_ENV" envDefault:"development"`
	MigrateOnBoot bool   `env:"MIGRATE_ON_BOOT" envDefault:"true"`

	HTTP   HTTP
	DB     DB
	Auth   Auth
	Logger Logger
	Prayer Prayer
}

// HTTP holds the HTTP server configuration.
type HTTP struct {
	Port            int           `env:"HTTP_PORT" envDefault:"8080"`
	ReadTimeout     time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"15s"`
	WriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"15s"`
	ShutdownTimeout time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT" envDefault:"10s"`
	// CORSAllowedOrigins lets the future SPA (different port) talk to the API.
	CORSAllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS" envSeparator:"," envDefault:"http://localhost:5173,http://localhost:3000"`
}

// DB holds PostgreSQL connection configuration.
type DB struct {
	URL      string `env:"DATABASE_URL" envDefault:"postgres://amana:amana@localhost:5432/amana?sslmode=disable"`
	MaxConns int32  `env:"DATABASE_MAX_CONNS" envDefault:"10"`
}

// Auth holds authentication configuration for the internal /api/app surface.
type Auth struct {
	JWTSecret string        `env:"JWT_SECRET" envDefault:"dev-secret-change-me"`
	JWTTTL    time.Duration `env:"JWT_TTL" envDefault:"24h"`
}

// Logger configures structured logging.
type Logger struct {
	Level  string `env:"LOG_LEVEL" envDefault:"info"`  // debug|info|warn|error
	Format string `env:"LOG_FORMAT" envDefault:"json"` // json|text
}

// Prayer holds the namaz calculation parameters (defaults: Grozny, Shafi'i).
type Prayer struct {
	Lat      float64 `env:"PRAYER_LAT" envDefault:"43.3178"`
	Lon      float64 `env:"PRAYER_LON" envDefault:"45.6949"`
	Madhab   string  `env:"PRAYER_MADHAB" envDefault:"shafii"`
	Timezone string  `env:"PRAYER_TIMEZONE" envDefault:"Europe/Moscow"`
}

// Load reads configuration from the process environment. A local .env file is
// loaded first if present (best-effort — absence is not an error).
func Load() (Config, error) {
	_ = godotenv.Load()

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
