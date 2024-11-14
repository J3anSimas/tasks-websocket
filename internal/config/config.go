package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBConnectionString string
	TokenSecret        string
}

var Cfg Config

func Instantiate() {
	godotenv.Load()
	Cfg.DBConnectionString = os.Getenv("DB_CONNECTION_STRING")
	Cfg.TokenSecret = os.Getenv("TOKEN_SECRET")
	if Cfg.DBConnectionString == "" {
		panic("DB_CONNECTION_STRING not found")
	}
	if Cfg.TokenSecret == "" {
		panic("TOKEN_SECRET not found")
	}
}
