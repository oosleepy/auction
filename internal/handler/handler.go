// Package handler deals with handler
package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"auction/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

type App struct {
	Redisconn *redis.Client
	M         sync.Mutex
	Cancel    context.CancelFunc
	Pgconn    *pgx.Conn
}

func (a *App) exitwriteroutine(ctx context.Context, expiry int) {
	select {
	case <-time.After(time.Duration(expiry)):
		bidAmt := a.Redisconn.Get(context.Background(), "Bid").Val()
		bidName := a.Redisconn.Get(context.Background(), "Name").Val()
		ip := a.Redisconn.Get(context.Background(), "IP").Val()
		_, err := a.Pgconn.Exec(context.Background(), "INSERT INTO bid_history (bid_name,bid_amt,ip) VALUES ($1,$2,$3) ", bidName, bidAmt, ip)
		if err != nil {
			log.Fatal(err)
		}
		a.Redisconn.Del(context.Background(), "Bid", "Name", "IP")
		a.Redisconn.Set(context.Background(), "Start", "ended", time.Duration(0))
	case <-ctx.Done():
		return
	}
}

func (a *App) Getbid(w http.ResponseWriter, r *http.Request) {
	var ResponseBid models.GetBid
	exist := a.Redisconn.Exists(context.Background(), "Bid").Val()
	endcheck := a.Redisconn.Get(context.Background(), "Start").Val()
	if endcheck == "ended" {
		http.Error(w, "Auction ended",http.StatusNotFound)
		return
	}else if exist == 0{
		http.Error(w, "Auction not started", http.StatusServiceUnavailable)
		return
	}else{
	bid := a.Redisconn.Get(context.Background(), "Bid").Val()
	name := a.Redisconn.Get(context.Background(), "Name").Val()
	ResponseBid.Bid = bid
	ResponseBid.Name = name
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&ResponseBid)
	}
}

func (a *App) Bid(w http.ResponseWriter, r *http.Request) {
	var UserBody models.UserBid
	json.NewDecoder(r.Body).Decode(&UserBody)
	userbid, err := strconv.Atoi(UserBody.UserBid)
	if err != nil {
		http.Error(w, "error conversion", http.StatusBadRequest)
		return
	}
	// mutexes locks this part for protecting race condition
	a.M.Lock()         // this is mutex format lock then defer then the code(critical section) to protect
	defer a.M.Unlock() // this solves teh issue of the 2 comments below written as thsi runs as soon as below retuns
	highestbid, err := strconv.Atoi(a.Redisconn.Get(context.Background(), "Bid").Val())
	if err != nil { // if its not nil -> no value really -> that happens when bid doesnt exist -> start/end
		endcheck := a.Redisconn.Get(context.Background(), "Start").Val()
		if endcheck == "ended" {
			http.Error(w, "Auction ended ", http.StatusNotFound)
			return // here i need ot set a m.Unlock to unlock this part if bid is not set
		}
		http.Error(w, "Auction not set", http.StatusServiceUnavailable)
		return

	}
	if userbid > highestbid {
		a.Redisconn.Set(context.Background(), "Bid", UserBody.UserBid, time.Duration(0))
		a.Redisconn.Set(context.Background(), "IP", r.RemoteAddr, time.Duration(0))
		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusConflict)
	}
	// m.Unlock() -> putting this alone here will not take care of the return if bid is not set yet
}

func (a *App) Setbid(w http.ResponseWriter, r *http.Request) {
	var setredis models.Setredis
	json.NewDecoder(r.Body).Decode(&setredis)	

	a.M.Lock()
	defer a.M.Unlock()
	if a.Cancel != nil{
		a.Cancel()
	}
	ctx,cancel := context.WithCancel(context.Background())
	a.Cancel = cancel

	expiry := setredis.Expiry * int(time.Second)
	a.Redisconn.Set(context.Background(), "Name", setredis.Name, time.Duration(0))
	a.Redisconn.Set(context.Background(), "Bid", setredis.Bid, time.Duration(0))
	a.Redisconn.Set(context.Background(), "IP", setredis.IP, time.Duration(0))
	a.Redisconn.Set(context.Background(), "Start", "active", time.Duration(0))

	go a.exitwriteroutine(ctx, expiry)
}
