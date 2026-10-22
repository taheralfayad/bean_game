package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DB    DBConfig
	Files FilesConfig
}

type FilesConfig struct {
	Endpoint       string
	Region         string
	Bucket         string
	AccessKey      string
	SecretKey      string
	PathStyle      bool
	PublicURL      string
	MaxUploadBytes int64
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
	pathStyle, err := strconv.ParseBool(envOr("S3_PATH_STYLE", "false"))
	if err != nil {
		return nil, fmt.Errorf("parse S3_PATH_STYLE: %w", err)
	}

	maxUploadBytes, err := strconv.ParseInt(envOr("UPLOAD_MAX_BYTES", "8388608"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse UPLOAD_MAX_BYTES: %w", err)
	}
	if maxUploadBytes <= 0 {
		return nil, fmt.Errorf("UPLOAD_MAX_BYTES must be positive")
	}

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
		Files: FilesConfig{
			Endpoint:  os.Getenv("S3_ENDPOINT"),
			Region:    os.Getenv("S3_REGION"),
			Bucket:    os.Getenv("S3_BUCKET"),
			AccessKey: os.Getenv("S3_ACCESS_KEY"),
			SecretKey: os.Getenv("S3_SECRET_KEY"),
			PathStyle: pathStyle,
			PublicURL: strings.TrimSuffix(os.Getenv("S3_PUBLIC_URL"), "/"),

			MaxUploadBytes: maxUploadBytes,
		},
	}

	var missing []string
	for name, value := range map[string]string{
		"POSTGRES_HOST":     cfg.DB.Host,
		"POSTGRES_USER":     cfg.DB.User,
		"POSTGRES_PASSWORD": cfg.DB.Password,
		"POSTGRES_DB":       cfg.DB.Name,
		"S3_REGION":         cfg.Files.Region,
		"S3_BUCKET":         cfg.Files.Bucket,
	} {
		if value == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	if (cfg.Files.AccessKey == "") != (cfg.Files.SecretKey == "") {
		return nil, fmt.Errorf("S3_ACCESS_KEY and S3_SECRET_KEY must be set together")
	}

	return cfg, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
