//Package models deals with models
package models



//go exports in capital but json serializes it so thats why the `json:"bid"`

type Setredis struct{
	Name string `json:"name"`
	Bid string `json:"bid"` 
	IP string		`json:"ip"`
	Expiry int `json:"time"`
}

type GetBid struct {
	Name string `json:"name"`
	Bid string `json:"bid"`

}

type UserBid struct{
	UserBid string `json:"userbid"`
}
