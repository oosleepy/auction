// Package handler deals with handler
package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"auction/internal/models"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

type App struct {
	Redisconn     *redis.Client
	M             map[string]*sync.Mutex
	Cancel        map[string]context.CancelFunc
	Pgconn        *pgx.Conn
	Metamutex     sync.Mutex
	Connectionmap map[string][]*websocket.Conn
	Realtimemutex map[string]*sync.Mutex
	MetaRealmutex sync.Mutex
}

func (a *App) exitwriteroutine(ctx context.Context, expiry int, namespace string) {
	select {
	case <-time.After(time.Duration(expiry)):
		bidAmt := a.Redisconn.Get(context.Background(), namespace+"bid").Val()
		bidName := a.Redisconn.Get(context.Background(), namespace+"name").Val()
		ip := a.Redisconn.Get(context.Background(), namespace+"ip").Val()
		_, err := a.Pgconn.Exec(context.Background(), "INSERT INTO bid_history (bid_name,bid_amt,ip) VALUES ($1,$2,$3) ", bidName, bidAmt, ip)
		if err != nil {
			log.Fatal(err)
		}
		a.Redisconn.Del(context.Background(), namespace+"bid", namespace+"name", namespace+"ip", namespace+"start")
	case <-ctx.Done():
		return
	}
}

var upgrader = websocket.Upgrader{}

func (a *App) Ws(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("name")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "failed to create conneciton", http.StatusBadGateway)
		return
	}
	a.MetaRealmutex.Lock()
	if a.Realtimemutex[query] == nil {
		a.Realtimemutex[query] = &sync.Mutex{}
	}
	a.MetaRealmutex.Unlock()

	a.Realtimemutex[query].Lock()
	defer a.Realtimemutex[query].Unlock()
	a.Connectionmap[query] = append(a.Connectionmap[query], conn)
}

func (a *App) Getbid(w http.ResponseWriter, r *http.Request) {
	var Response models.GetBidResponse
	var Request models.GetBidRequest
	json.NewDecoder(r.Body).Decode(&Request)
	namespace := "auction:" + Request.Name + ":"
	endcheck := a.Redisconn.Get(context.Background(), namespace+"start").Val()
	if endcheck == "active" {
		bid := a.Redisconn.Get(context.Background(), namespace+"bid").Val()
		name := a.Redisconn.Get(context.Background(), namespace+"name").Val()
		Response.Bid = bid
		Response.Name = name
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&Response)
	} else {
		var name string
		err := a.Pgconn.QueryRow(context.Background(), "SELECT bid_name FROM bid_history WHERE bid_name = $1", Request.Name).Scan(&name)
		if err != nil {
			http.Error(w, "Auction not started", http.StatusServiceUnavailable)
			return
		} else {
			http.Error(w, "Auction ended", http.StatusNotFound)
			return
		}
	}
}

func (a *App) Bid(w http.ResponseWriter, r *http.Request) {
	var UserBody models.UserBidRequest
	json.NewDecoder(r.Body).Decode(&UserBody)
	userbid, err := strconv.Atoi(UserBody.Bid)
	if err != nil {
		http.Error(w, "error conversion", http.StatusBadRequest)
		return
	}
	namespace := "auction:" + UserBody.Name + ":"

	a.Metamutex.Lock()
	if a.M[UserBody.Name] == nil {
		a.M[UserBody.Name] = &sync.Mutex{}
	}
	a.Metamutex.Unlock()
	a.M[UserBody.Name].Lock()
	defer a.M[UserBody.Name].Unlock()
	highestbid, err := strconv.Atoi(a.Redisconn.Get(context.Background(), namespace+"bid").Val())
	if err != nil { // if its not nil -> no value really -> that happens when bid doesnt exist -> start/end
			http.Error(w, "Auction ended/not started ", http.StatusNotFound)
			return
	}
	if userbid > highestbid {
		a.Redisconn.Set(context.Background(), namespace+"bid", UserBody.Bid, time.Duration(0))
		a.Redisconn.Set(context.Background(), namespace+"ip", r.RemoteAddr, time.Duration(0))

		wsReponse := models.GetBidResponse{
			Name: UserBody.Name,
			Bid:  UserBody.Bid,
		}
		a.MetaRealmutex.Lock()
		if a.Realtimemutex[UserBody.Name] == nil {
			a.Realtimemutex[UserBody.Name] = &sync.Mutex{}
		}
		a.MetaRealmutex.Unlock()
		a.Realtimemutex[UserBody.Name].Lock()
		defer a.Realtimemutex[UserBody.Name].Unlock()
		for _, conn := range a.Connectionmap[UserBody.Name] {
			conn.WriteJSON(wsReponse)
		}
		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusConflict)
	}
}

func (a *App) Setbid(w http.ResponseWriter, r *http.Request) {
	var setredis models.Setredis
	json.NewDecoder(r.Body).Decode(&setredis)

	a.Metamutex.Lock()
	if a.M[setredis.Name] == nil {
		a.M[setredis.Name] = &sync.Mutex{}
	}
	a.Metamutex.Unlock()

	a.M[setredis.Name].Lock()
	defer a.M[setredis.Name].Unlock()
	if a.Cancel[setredis.Name] != nil {
		a.Cancel[setredis.Name]()
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.Cancel[setredis.Name] = cancel

	expiry := setredis.Expiry * int(time.Second)
	namespace := "auction:" + setredis.Name + ":"
	a.Redisconn.Set(context.Background(), namespace+"name", setredis.Name, time.Duration(0))
	a.Redisconn.Set(context.Background(), namespace+"bid", setredis.Bid, time.Duration(0))
	a.Redisconn.Set(context.Background(), namespace+"ip", setredis.IP, time.Duration(0))
	a.Redisconn.Set(context.Background(), namespace+"start", "active", time.Duration(0))

	go a.exitwriteroutine(ctx, expiry, namespace)
}



func (a *App) Listactive(w http.ResponseWriter, r  *http.Request){
	auctionlist := a.Redisconn.Keys(context.Background(), "auction:*:start").Val()
	var namelist []string
	
	for _,key := range(auctionlist){
		parts:=strings.Split(key, ":")
		namelist = append(namelist, parts[1])
	}
	// if exist thenshow if not it become empty
	activelist := models.ActiveListResponse{
		ActiveList: namelist,
	}	
	json.NewEncoder(w).Encode(activelist)
}

func (a *App) Listhistory(w http.ResponseWriter, r *http.Request){
	auctionlist,err := a.Pgconn.Query(context.Background(), "SELECT bid_name,bid_amt,ip,created_at FROM bid_history")
	if err!= nil{
		http.Error(w, "No history", http.StatusNotFound)
		return
	}

	var perrowstruct models.ListHistoryPerRow	
	var history models.ListHistoryResponse
	for auctionlist.Next(){
		err := auctionlist.Scan(&perrowstruct.Bidname, &perrowstruct.Bidamt, &perrowstruct.IP, &perrowstruct.Createdat)
		if err!=nil{
			fmt.Println(err)
		}
		history = append(history, perrowstruct)
	}

	json.NewEncoder(w).Encode(history)
}
