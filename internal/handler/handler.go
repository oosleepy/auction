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
	Mutexbid      map[string]*sync.Mutex   // common and used by both /bid and /setbid to control race conditions for 3 cases -> (/bid , bid)(/setbid ,/setbid) (/bit , /setbid)
	Cancel        map[string]context.CancelFunc
	Pgconn        *pgx.Conn
	Metamutex     sync.Mutex  //meta check if mutexbid is already exists or not
	Connectionmap map[string][]*websocket.Conn   //used to create a list of connection for a certain auction for websocket ie [auctioname] = [coonection1,connection2....]
	Realtimemutex map[string]*sync.Mutex //common and used by both /ws and /bid to control the write and read of the connectionmap array without panic
	MetaRealmutex sync.Mutex
}

func (a *App) exitwriteroutine(ctx context.Context, expiry int, namespace string) {
	select {

	//CASE 1 -> Auction ended or expiry is done 
	case <-time.After(time.Duration(expiry)):
		bidAmt := a.Redisconn.Get(context.Background(), namespace+"bid").Val()
		bidName := a.Redisconn.Get(context.Background(), namespace+"name").Val()
		ip := a.Redisconn.Get(context.Background(), namespace+"ip").Val()
		_, err := a.Pgconn.Exec(context.Background(), "INSERT INTO bid_history (bid_name,bidamt,ip) VALUES ($1,$2,$3) ", bidName, bidAmt, ip)
		if err != nil {
			log.Fatal(err)
		}
		a.Redisconn.Del(context.Background(), namespace+"bid", namespace+"name", namespace+"ip", namespace+"start")

	//Case 2 -> If cancel function is called ie the already running so overwrite case then this happens
	case <-ctx.Done():
		return
	}
}

var upgrader = websocket.Upgrader{}

//Ws fucntion just creates the connection and appends it thats it no writing to display
func (a *App) Ws(w http.ResponseWriter, r *http.Request) {

	queryauction := r.URL.Query().Get("name")  //request's url's query that is name=?

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "failed to create conneciton", http.StatusBadGateway)
		return
	}

	a.MetaRealmutex.Lock()
	if a.Realtimemutex[queryauction] == nil {
		a.Realtimemutex[queryauction] = &sync.Mutex{}
	}
	a.MetaRealmutex.Unlock()
	
	//cuz writing into array is always a race condition we use mutex to govern it
	a.Realtimemutex[queryauction].Lock()
	defer a.Realtimemutex[queryauction].Unlock()
	// appending connection to an auction ie [auctionname] = [conenction1 , connection2 .... ]
	a.Connectionmap[queryauction] = append(a.Connectionmap[queryauction], conn)
}


func (a *App) Getbid(w http.ResponseWriter, r *http.Request) {
	var Response models.GetBidResponse
	var Request models.GetBidRequest
	json.NewDecoder(r.Body).Decode(&Request)

	namespace := "auction:" + Request.Name + ":"  //this is done for redis so that we can idetify for which auction in redis
	endcheck := a.Redisconn.Get(context.Background(), namespace+"start").Val()

	//3 case -> active , not started , ended
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
	userbid, err := strconv.Atoi(UserBody.Bid) // converting string to int
	if err != nil { //if error shows bad request 
		http.Error(w, "error conversion", http.StatusBadRequest)
		return
	}
	namespace := "auction:" + UserBody.Name + ":"
	
	//meta lock for checing if a mutex for the auction already exists or not if not we create one
	a.Metamutex.Lock()
	if a.Mutexbid[UserBody.Name] == nil {
		a.Mutexbid[UserBody.Name] = &sync.Mutex{}
	}
	a.Metamutex.Unlock()


	a.Mutexbid[UserBody.Name].Lock()
	defer a.Mutexbid[UserBody.Name].Unlock()
	highestbid, err := strconv.Atoi(a.Redisconn.Get(context.Background(), namespace+"bid").Val())
	if err != nil { // if its not nil -> no value really -> that happens when bid doesnt exist -> start/end
			http.Error(w, "Auction ended/not started ", http.StatusNotFound)
			return
	}
	if userbid > highestbid {
		a.Redisconn.Set(context.Background(), namespace+"bid", UserBody.Bid, time.Duration(0))
		a.Redisconn.Set(context.Background(), namespace+"ip", r.RemoteAddr, time.Duration(0))
		

		//make changes to the websocker connection of this auction
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

		//go thru the connection map list of this aucciton name ie [auctioname] = [connection1,connection2,....]
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
	
	
	//this mutex is to initialize our set biding mutex 
	//liek if ak47 bid is being set then ->
	//check mutex[ak47] exist if not then make mutex thats what this below mutex does
	a.Metamutex.Lock()
	if a.Mutexbid[setredis.Name] == nil {
		a.Mutexbid[setredis.Name] = &sync.Mutex{}
	}
	a.Metamutex.Unlock()
	
	//using mutex to lock the setting of bids to bypass or like a fix/remedy for  race condition
	a.Mutexbid[setredis.Name].Lock()
	defer a.Mutexbid[setredis.Name].Unlock()  //defer makes it such that the following has to execute then the defer line 
	
	//Cancel function is implemented here so that using the context the bid created would be deleted from redis after the expiry time
	
	//check if cancel function for the auction exist if it does exist then cancel it 
	//this case is like when i already have a ak47 auction runnign and i create another a47 auction which i overwrite for this i need to stop the previous expirty that is this
	if a.Cancel[setredis.Name] != nil {
		a.Cancel[setredis.Name]()
	}
	
	//create a ctx and cancel function 
	ctx, cancel := context.WithCancel(context.Background())
	a.Cancel[setredis.Name] = cancel  //put cancel function in [auctionname] = CancelFunc
	

	//namespace is used to identify which auction is which its bascially auction:name:key
	expiry := setredis.Expiry * int(time.Second)
	namespace := "auction:" + setredis.Name + ":"
	a.Redisconn.Set(context.Background(), namespace+"name", setredis.Name, time.Duration(0))
	a.Redisconn.Set(context.Background(), namespace+"bid", setredis.Bid, time.Duration(0))
	a.Redisconn.Set(context.Background(), namespace+"ip", setredis.IP, time.Duration(0))
	a.Redisconn.Set(context.Background(), namespace+"start", "active", time.Duration(0))
	
	//Routine to end the auction after expirty -> by using ctx's cancel 
	//ctx is send to go routine to stop the routine/func if cacnel funciton for the already running overwrite case is found or called 
	go a.exitwriteroutine(ctx, expiry, namespace)
}



func (a *App) Listactive(w http.ResponseWriter, r  *http.Request){
	
	//creates a list of auction names that are starting with auction and ending with start -> this gives all auction names that is active
	auctionlist := a.Redisconn.Keys(context.Background(), "auction:*:start").Val()
	var namelist []string
	
	for _,key := range(auctionlist){
		parts:=strings.Split(key, ":")  //parts -> [auction,name,start]
		namelist = append(namelist, parts[1])  // namelist = [name1.name2....]
	}
	// if exist thenshow if not it become empty
	activelist := models.ActiveListResponse{
		ActiveList: namelist,
	}

	json.NewEncoder(w).Encode(activelist)
}

func (a *App) Listhistory(w http.ResponseWriter, r *http.Request){
	auctionlist,err := a.Pgconn.Query(context.Background(), "SELECT bid_name,bidamt,ip,created_at FROM bid_history")
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
