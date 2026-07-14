// Package db deal with db
package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Pgconn(pgurl string) (*pgxpool.Pool,error) {
	conn,err := pgxpool.New(context.Background(),pgurl)
	if err != nil{
		return nil,err
	}
	return conn,nil
}
