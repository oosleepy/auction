package main

import (
	"log"
	"net/http"

	"auction/config"
	redis "auction/internal/cache"
	"auction/internal/handler"

	"github.com/joho/godotenv"
)

func main() {

	godotenv.Load()
	cfg := config.Loadconfig()
	conn, err := redis.Redisconn(cfg.RedisURL)
	if err!= nil{
		log.Fatal("redis connection error")
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/setbid", handler.Setbid(conn))
	mux.HandleFunc("/bid", handler.Bid(conn))
	mux.HandleFunc("/getbid", handler.Getbid(conn))

	log.Fatal(http.ListenAndServe(":8080", mux))
}
