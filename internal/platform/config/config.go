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

	HTTP      HTTP
	DB        DB
	Auth      Auth
	Logger    Logger
	Prayer    Prayer
	Financing Financing
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

// Financing holds business settings for the murabaha core.
type Financing struct {
	// ComparisonRatePercent is the illustrative annual rate used only for the
	// "no-riba vs conventional credit" comparison shown in the contract wizard.
	ComparisonRatePercent int `env:"FINANCING_COMPARISON_RATE_PERCENT" envDefault:"28"`
}

// Prayer holds the namaz calculation and scheduling parameters (defaults:
// Grozny, Shafi'i, MWL; 20-minute buffer after each prayer; Jummah 12:30–14:00).
type Prayer struct {
	Lat             float64 `env:"PRAYER_LAT" envDefault:"43.3178"`
	Lon             float64 `env:"PRAYER_LON" envDefault:"45.6949"`
	Madhab          string  `env:"PRAYER_MADHAB" envDefault:"shafii"`
	Timezone        string  `env:"PRAYER_TIMEZONE" envDefault:"Europe/Moscow"`
	Method          string  `env:"PRAYER_METHOD" envDefault:"MWL"`
	BufferBeforeMin int     `env:"PRAYER_BUFFER_BEFORE_MIN" envDefault:"0"`
	BufferAfterMin  int     `env:"PRAYER_BUFFER_AFTER_MIN" envDefault:"20"`
	JummahStart     string  `env:"PRAYER_JUMMAH_START" envDefault:"12:30"`
	JummahEnd       string  `env:"PRAYER_JUMMAH_END" envDefault:"14:00"`
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
