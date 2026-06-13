//Package config deals with config 
package config

import(
	"os"
)

type Config struct{
	RedisURL string
}

func Loadconfig() Config{
	return Config{
		RedisURL: os.Getenv("REDIS_URL"),
	}
}
