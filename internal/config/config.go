package config

import (
	"os"
	"strings"
)

const DefaultProjectName = "default"

type Config struct {
	HTTPAddr       string
	GRPCAddr       string
	DatabaseDriver string
	DatabaseDSN    string
	DefaultProject string
}

func Load() Config {
	dsn := getenv("AGENTTRACE_DATABASE_DSN", "")
	driver := getenv("AGENTTRACE_DATABASE_DRIVER", "")
	if databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL")); databaseURL != "" && dsn == "" {
		dsn = databaseURL
	}
	if driver == "" {
		driver = inferDriver(dsn)
	}
	if dsn == "" {
		driver = "sqlite"
		dsn = getenv("AGENTTRACE_SQLITE_PATH", "data/agenttrace.db")
	}
	return Config{
		HTTPAddr:       getenv("AGENTTRACE_HTTP_ADDR", ":6006"),
		GRPCAddr:       getenv("AGENTTRACE_GRPC_ADDR", ":4317"),
		DatabaseDriver: driver,
		DatabaseDSN:    dsn,
		DefaultProject: getenv("AGENTTRACE_DEFAULT_PROJECT", DefaultProjectName),
	}
}

func inferDriver(dsn string) string {
	lower := strings.ToLower(strings.TrimSpace(dsn))
	switch {
	case strings.HasPrefix(lower, "postgres://"), strings.HasPrefix(lower, "postgresql://"), strings.Contains(lower, "host="):
		return "postgres"
	default:
		return "sqlite"
	}
}

func getenv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
