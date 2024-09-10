package config

import (
	"flag"
	"os"
)

// Config Структура со всеми конфигурациями сервера
type Config struct {
	Addr       string
	DB         string
	SecretKey  string
	AccrualAdr string
}

// MustLoadConfig загрузка конфигурации
func MustLoadConfig() *Config {
	var config Config
	config.SecretKey = "22gwiT5#eQxdh89OJZM-9af=LDB^EIJsW7Bbv90s1L^U.O7jNu8OrEhWLM.zJFUk"
	flag.StringVar(&config.Addr, "a", "localhost:8081", "Адрес порт сервера")
	flag.StringVar(
		&config.DB,
		"d",
		"user=postgres password=postgres host=localhost port=5432 database=postgres sslmode=disable pool_max_conns=10",
		"DSN строка для соединения с базой данных",
	)
	flag.StringVar(
		&config.AccrualAdr,
		"r",
		"http://localhost:8080",
		"Адрес системы расчёта начислений",
	)
	flag.Parse()

	envAddr := os.Getenv("RUN_ADDRESS")
	if envAddr != "" {
		config.Addr = envAddr
	}
	envDB := os.Getenv("DATABASE_URI")
	if envDB != "" {
		config.DB = envDB
	}
	envAccrualAdr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS")
	if envAccrualAdr != "" {
		config.AccrualAdr = envAccrualAdr
	}

	return &config
}
