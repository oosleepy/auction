//Package redis deals with redis
package redis

import(
	"github.com/redis/go-redis/v9"
)

func Redisconn(redisurl string ) (*redis.Client,error) {
	conn, err := redis.ParseURL(redisurl)
	if err != nil{
		return nil, err
	}
	client := redis.NewClient(conn)
	return client,nil
}


