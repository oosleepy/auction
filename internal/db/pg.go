//Package db deal with db
package db

import(
	"github.com/jackc/pgx/v5"
	"context"
)

func Pgconn(pgurl string) (*pgx.Conn,error) {
	conn,err := pgx.Connect(context.Background(),pgurl)
	if err != nil{
		return nil,err
	}
	return conn,nil
}
