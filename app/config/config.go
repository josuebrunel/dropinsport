package config

import (
	"os"
	"strconv"
)

type Config struct {
	HTTPAddr string
	DBHost   string
	DBPort   int
	DBUser   string
	DBPass   string
	DBName   string
}

func getValue(envKey string) string {
	var val string
	if v := os.Getenv(envKey); v != "" {
		val = v
	}
	return val
}

func NewConfig() Config {
	var cfg Config
	cfg.HTTPAddr = ":8888"
	if v := getValue("SDI_HTTP"); v != "" {
		cfg.HTTPAddr = v
	}
	cfg.DBHost = getValue("SDI_DB_HOST")
	var port = 5432
	if p, err := strconv.Atoi(getValue("SDI_DB_HOST")); err == nil {
		port = p
	}
	cfg.DBPort = port
	cfg.DBUser = getValue("SDI_DB_USERNAME")
	cfg.DBPass = getValue("SDI_DB_PASSWORD")
	cfg.DBName = getValue("SDI_DB_NAME")
	return cfg
}
