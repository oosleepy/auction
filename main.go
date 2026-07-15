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
	"auction/internal/auth"
	"github.com/joho/godotenv"
	"github.com/gorilla/websocket"
	
)

func main() {

	
	_ = godotenv.Load() //loads the env variables -> redis url , postgres url
	cfg := config.Loadconfig() //returns a struct with attr redisurl and postgres url from env
	redisconn, err := redis.Redisconn(cfg.RedisURL)
	if err!= nil{
		log.Fatal("redis connection error")
	}
	pgconn,err := db.Pgconn(cfg.PostgresURL)
	if err != nil{
		log.Fatal("postgres connection error")
	}
	
	//without make -> creates a nil map which when writing -> panic
	//with make -> underlying mememory is handled -> when writing -> no panic
	cancel := make(map[string]context.CancelFunc) //create a hashmap [string] = cancelFunc
	m := make(map[string]*sync.Mutex) //creates a hashmap [string] = *sync.Mutex
	connectionmap := make(map[string][]*websocket.Conn) //creates hashmap [string] = [*websocket.Con,*websoccekt.con....]
	realtimemutex := make(map[string]*sync.Mutex) // creates hashmap [string] = *sync.Mutex
	
	//create a struct with methods which all handler are functions of , yk?
	//we use struct with methods to persist the memory of the redisconn and postgresconn mutexes etc across several handler calls
	//if struct with methods was not used then each time handler is called a separate redis and postgres and mutex is called ykyk
	app := &handler.App{
		Redisconn: redisconn,
		Pgconn:pgconn,
		Mutexbid:m,
		Cancel:cancel,
		Connectionmap: connectionmap,
		Realtimemutex: realtimemutex,
	}

	
	mux := http.NewServeMux()	
	mux.HandleFunc("/setbid", auth.VerifyMiddleware(http.HandlerFunc(app.Setbid)))
	mux.HandleFunc("/bid", auth.VerifyMiddleware(http.HandlerFunc(app.Bid)))
	mux.HandleFunc("/getbid", app.Getbid)
	mux.HandleFunc("/ws", app.Ws)
	mux.HandleFunc("/listactive", app.Listactive)
	mux.HandleFunc("/listhistory", app.Listhistory)
	mux.HandleFunc("/register", app.Register)
	mux.HandleFunc("/login",app.Login)
	mux.HandleFunc("/mybid",auth.VerifyMiddleware(http.HandlerFunc(app.GetTotalbid)))

	fs := http.FileServer(http.Dir("./frontend"))
	mux.Handle("/", fs)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
