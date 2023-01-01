package env

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strings"
)

func envString(key string) string {
	s, ok := os.LookupEnv(key)
	if !ok {
		log.Panicf("[env] Cannot find var `%s`", key)
	}
	return s
}

var (
	port   string
	domain string
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Print("[env] Cannot find .env file")
	} else {
		log.Print("[env] .env file loaded")
	}

	port = strings.TrimLeft(envString(`PORT`), ":")
	domain = envString("DOMAIN")
}

func Port() string {
	return port
}

func Domain() string {
	return domain
}
