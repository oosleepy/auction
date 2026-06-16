package main

import (
	"log"
	"net/http"

	"auction/config"
	redis "auction/internal/cache"
	"auction/internal/handler"
	"auction/internal/db"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	cfg := config.Loadconfig()
	redisconn, err := redis.Redisconn(cfg.RedisURL)
	if err!= nil{
		log.Fatal("redis connection error")
	}
	pgconn,err := db.Pgconn(cfg.PostgresURL)
	if err != nil{
		log.Fatal("postgres connection error")
	}	
	app := &handler.App{
		Redisconn: redisconn,
		Pgconn:pgconn,	
	}
	mux := http.NewServeMux()	
	mux.HandleFunc("/setbid", app.Setbid)
	mux.HandleFunc("/bid", app.Bid)
	mux.HandleFunc("/getbid", app.Getbid)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
