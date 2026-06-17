//Package models deals with models
package models



//go exports in capital but json serializes it so thats why the `json:"bid"`

type Setredis struct{
	Name string `json:"name"`
	Bid string `json:"bid"` 
	IP string		`json:"ip"`
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


