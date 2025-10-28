package config

import (
	"fmt"
	"os"
)

type Config struct {
	DBHost string
	DBPort string
	DBName string
	DBUser string
	DBPass string
	DBSSL  string

	HTTPAddr string

	StanClusterID string
	StanClientID  string
	StanURL       string
	Channel       string
	Durable       string
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func Load() Config {
	return Config{
		DBHost: getenv("DB_HOST", "localhost"),
		DBPort: getenv("DB_PORT", "5432"),
		DBName: getenv("DB_NAME", "orders"),
		DBUser: getenv("DB_USER", "orders"),
		DBPass: getenv("DB_PASS", "orders"),
		DBSSL:  getenv("DB_SSLMODE", "disable"),

		HTTPAddr: getenv("HTTP_ADDR", ":8080"),

		StanClusterID: getenv("STAN_CLUSTER_ID", "test-cluster"),
		StanClientID:  getenv("STAN_CLIENT_ID", "orders-server-1"),
		StanURL:       getenv("STAN_URL", "nats://localhost:4222"),
		Channel:       getenv("CHANNEL", "orders"),
		Durable:       getenv("DURABLE", "orders-durable"),
	}
}

func (c Config) PGConnString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", c.DBUser, c.DBPass, c.DBHost, c.DBPort, c.DBName, c.DBSSL)
}
