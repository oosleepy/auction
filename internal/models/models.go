// Package models deals with models
package models

import "time"

//go exports in capital but json serializes it so thats why the `json:"bid"`

type Setredis struct{
	Name string `json:"name"`
	Bid string `json:"bid"` 
	Expiry int `json:"time"`
}

 
type GetBidResponse struct {
	Name string `json:"name"`
	Bid string `json:"bid"`

}
type GetBidRequest struct {
	Name string `json:"name"`
}

type UserBidRequest struct{
	Bid string `json:"bid"`
	Name string `json:"name"`
}

type ActiveListResponse struct{
	ActiveList []string `json:"active_auciton"`
}

type ListHistoryPerRow struct{
	Bidname string `json:"bid_name"`
	Bidamt string `json:"bid_amt"`
	UserID string `json:"userid"`
	Createdat time.Time `json:"created_at"`
}

type ListHistoryResponse []ListHistoryPerRow

type UserRequest struct{
	Username string `json:"username"`
	Password string `json:"password"`
}
