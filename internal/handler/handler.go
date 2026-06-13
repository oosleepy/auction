// Package handler deals with handler
package handler

import (
	"auction/internal/models"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func Getbid(redisconn *redis.Client) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		var ResponseBid models.GetBid
		bid := redisconn.Get(context.Background(), "Bid")
		ResponseBid.Bid = bid.Val()
		w.Header().Set("Content-Type","application/json")
		json.NewEncoder(w).Encode(&ResponseBid)
	}
}


func Bid(redisconn *redis.Client) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		var UserBody models.UserBid
		json.NewDecoder(r.Body).Decode(&UserBody)
		highestbid,err := strconv.Atoi(redisconn.Get(context.Background(),"Bid").Val())
		if err!=nil{
			http.Error(w,"Auction not set",http.StatusServiceUnavailable)
			return
		}
		userbid,err := strconv.Atoi(UserBody.UserBid)
		if err != nil{
			http.Error(w,"error conversion",http.StatusBadRequest)
			return
		}
		if userbid > highestbid{
			expire := redisconn.ExpireTime(context.Background(), "Bid").Val()
			redisconn.Set(context.Background(), "Bid", UserBody.UserBid,expire)
			redisconn.Set(context.Background(), "IP", r.RemoteAddr,expire)
			w.WriteHeader(http.StatusAccepted)
		}else{
			w.WriteHeader(http.StatusConflict)	
		}
	}
}


func Setbid(redisconn *redis.Client) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		var setredis models.Setredis

		json.NewDecoder(r.Body).Decode(&setredis)
		expiry := setredis.Expiry * int(time.Second)	
		redisconn.Set(context.Background(), "Bid",setredis.Bid,time.Duration(expiry))
		redisconn.Set(context.Background(), "IP",setredis.IP,time.Duration(expiry))

	
	}
}
