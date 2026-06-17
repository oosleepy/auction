package main

import (
	"context"
	"log"
	"net/http"
	"sync"

	"auction/config"
	redis "auction/internal/cache"
	"auction/internal/db"
	"auction/internal/handler"

	"github.com/joho/godotenv"
	"github.com/gorilla/websocket"
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
	cancel := make(map[string]context.CancelFunc)
	m := make(map[string]*sync.Mutex)
	connectionmap := make(map[string][]*websocket.Conn)
	realtimemutex := make(map[string]*sync.Mutex)
	app := &handler.App{
		Redisconn: redisconn,
		Pgconn:pgconn,
		M:m,
		Cancel:cancel,
		Connectionmap: connectionmap,
		Realtimemutex: realtimemutex,
	}
	mux := http.NewServeMux()	
	mux.HandleFunc("/setbid", app.Setbid)
	mux.HandleFunc("/bid", app.Bid)
	mux.HandleFunc("/getbid", app.Getbid)
	mux.HandleFunc("/ws", app.Ws)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
