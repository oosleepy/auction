package handler

import (
	"auction/config"
	redis "auction/internal/cache"
	"auction/internal/db"
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

func setup() *App {
	godotenv.Load("../../.env")
	cfg := config.Loadconfig()
	redisconn,err := redis.Redisconn(cfg.RedisURL)
	if err!=nil{
		log.Fatal("TESET redis connections error")
	}

	pgconn,err := db.Pgconn(cfg.PostgresURL)
	if err!=nil{
		log.Fatal("TEST postgres connecton error")
	}

	mutexbid := make(map[string]*sync.Mutex)
	cancel := make(map[string]context.CancelFunc)
	connectionmap := make(map[string][]*websocket.Conn)
	realtimemutex := make(map[string]*sync.Mutex)
	
	app := &App{
		Pgconn: pgconn,
		Redisconn: redisconn,
		Mutexbid: mutexbid,
		Cancel: cancel,
		Connectionmap: connectionmap,
		Realtimemutex: realtimemutex,
	}
	return app
}

func TestRaceConditons(t *testing.T){
	results := make(chan int, 2)
	app := setup()
	var wg sync.WaitGroup
	app.Redisconn.Set(context.Background(), "auction:ak47:bid", "10", time.Duration(0))
	app.Redisconn.Set(context.Background(), "auction:ak47:winner", "10", time.Duration(0))
	app.Redisconn.Set(context.Background(), "auction:ak47:start", "active", time.Duration(0))
	wg.Add(2)

	go func(){

		defer wg.Done()	
		body := bytes.NewBufferString(`{"bid":"200","name":"ak47"}`)
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/bid",body)
		ctx := context.WithValue(r.Context(), "userid", 1)
		r = r.WithContext(ctx)
		app.Bid(w,r)
		results <- w.Result().StatusCode
	}()

		go func(){

		defer wg.Done()	
		body := bytes.NewBufferString(`{"bid":"200","name":"ak47"}`)
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/bid",body)
		ctx := context.WithValue(r.Context(), "userid", 1)
		r = r.WithContext(ctx)
		app.Bid(w,r)
		results <- w.Result().StatusCode
	}()

	wg.Wait()
	close(results)

	codes := []int{}
	for code := range results {
    codes = append(codes, code)

	}
	got202 := codes[0] == http.StatusAccepted || codes[1] == http.StatusAccepted
	got409 := codes[0] == http.StatusConflict || codes[1] == http.StatusConflict

	if !got202 || !got409 {
		t.Errorf("expected one 202 and one 409, got %d and %d", codes[0], codes[1])
	}
}




