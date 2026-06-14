// Package handler deals with handler
package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"auction/internal/models"

	"github.com/redis/go-redis/v9"
)

func Getbid(redisconn *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ResponseBid models.GetBid
		exist := redisconn.Exists(context.Background(), "Bid").Val()
		endcheck := redisconn.Get(context.Background(), "Start").Val()
		if exist == 0 {
			if endcheck == "true" { // end state
				http.Error(w, "Auction ended", http.StatusNotFound)
				return
			}
			http.Error(w, "Auction not started/ set", http.StatusServiceUnavailable)
			return
		}
		bid := redisconn.Get(context.Background(), "Bid").Val()
		ResponseBid.Bid = bid
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&ResponseBid)
	}
}

// end check 3 States
// bid exist and nto empty -> auction going on
// bid no exist -> can be auction end/ auction start
// thats why we need start


func Bid(redisconn *redis.Client, m *sync.Mutex) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var UserBody models.UserBid
		json.NewDecoder(r.Body).Decode(&UserBody)
		userbid, err := strconv.Atoi(UserBody.UserBid)
		if err != nil {
			http.Error(w, "error conversion", http.StatusBadRequest)
			return
		}
		// mutexes locks this part for protecting race condition
		m.Lock()         // this is mutex format lock then defer then the code(critical section) to protect
		defer m.Unlock() // this solves teh issue of the 2 comments below written as thsi runs as soon as below retuns
		highestbid, err := strconv.Atoi(redisconn.Get(context.Background(), "Bid").Val())
		if err != nil {
			endcheck := redisconn.Get(context.Background(), "Start").Val()
			if endcheck == "true" {
				http.Error(w, "Auction ended ", http.StatusNotFound)
				return // here i need ot set a m.Unlock to unlock this part if bid is not set
			}
			http.Error(w, "Auction not set", http.StatusServiceUnavailable)
			return

		}
		if userbid > highestbid {
			expire := redisconn.ExpireTime(context.Background(), "Bid").Val()
			redisconn.Set(context.Background(), "Bid", UserBody.UserBid, expire)
			redisconn.Set(context.Background(), "IP", r.RemoteAddr, expire)
			w.WriteHeader(http.StatusAccepted)
		} else {
			w.WriteHeader(http.StatusConflict)
		}
		// m.Unlock() -> putting this alone here will not take care of the return if bid is not set yet
	}
}

func Setbid(redisconn *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var setredis models.Setredis

		json.NewDecoder(r.Body).Decode(&setredis)
		expiry := setredis.Expiry * int(time.Second)
		redisconn.Set(context.Background(), "Bid", setredis.Bid, time.Duration(expiry))
		redisconn.Set(context.Background(), "IP", setredis.IP, time.Duration(expiry))
		redisconn.Set(context.Background(), "Start", "true", time.Duration(0))
	}
}
