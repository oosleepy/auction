//Package config deals with config 
package config

import(
	"os"
)

type Config struct{
	RedisURL string
	PostgresURL string
}

func Loadconfig() Config{
	return Config{
		RedisURL: os.Getenv("REDIS_URL"),
		PostgresURL: os.Getenv("POSTGRES_URL"),
	}
}
