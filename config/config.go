//Package config returns a config struct with redisurl and posgresurl attributes 
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
