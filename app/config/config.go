package config

import (
	"fmt"
	"os"
)

type Config struct {
	HTTPAddr string
	DBHost   string
	DBPort   string
	DBUser   string
	DBPass   string
	DBName   string
}

func (cfg Config) GetDBDSN() string {
	dsn := "host=%s port=%s dbname=%s user='%s' password=%s sslmode=disable"
	return fmt.Sprintf(dsn, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBUser, cfg.DBPass)
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
	cfg.DBPort = "5432"
	if p := getValue("SDI_DB_PORT"); p != "" {
		cfg.DBPort = p
	}
	cfg.DBUser = getValue("SDI_DB_USERNAME")
	cfg.DBPass = getValue("SDI_DB_PASSWORD")
	cfg.DBName = getValue("SDI_DB_NAME")
	return cfg
}
