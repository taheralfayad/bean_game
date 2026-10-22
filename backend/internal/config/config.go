package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

type Config struct {
	DB DBConfig
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string

	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func (c DBConfig) DSN() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   net.JoinHostPort(c.Host, c.Port),
		Path:   "/" + c.Name,
	}
	q := url.Values{}
	q.Set("sslmode", c.SSLMode)
	u.RawQuery = q.Encode()
	return u.String()
}

func Load() (*Config, error) {
	cfg := &Config{
		DB: DBConfig{
			Host:     os.Getenv("POSTGRES_HOST"),
			Port:     envOr("POSTGRES_PORT", "5432"),
			User:     os.Getenv("POSTGRES_USER"),
			Password: os.Getenv("POSTGRES_PASSWORD"),
			Name:     os.Getenv("POSTGRES_DB"),
			SSLMode:  envOr("POSTGRES_SSLMODE", "disable"),

			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 30 * time.Minute,
		},
	}

	var missing []string
	for name, value := range map[string]string{
		"POSTGRES_HOST":     cfg.DB.Host,
		"POSTGRES_USER":     cfg.DB.User,
		"POSTGRES_PASSWORD": cfg.DB.Password,
		"POSTGRES_DB":       cfg.DB.Name,
	} {
		if value == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
