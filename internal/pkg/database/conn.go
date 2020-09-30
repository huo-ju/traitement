package database

import (
	"context"
    "time"
    "fmt"
	"github.com/jackc/pgx/v4/pgxpool"
)

var DBConn *Db

type Db struct {
	pool *pgxpool.Pool
}

func New(connstr string) (*Db, error) {
	db := new(Db)
	poolConfig, err := pgxpool.ParseConfig(connstr)
	if err != nil {
		return nil, err
	}
	poolConfig.MaxConns = 6
	db.pool, err = pgxpool.ConnectConfig(context.Background(), poolConfig)
    for err!=nil {
        fmt.Println(err)
        fmt.Println("wait 5 Second for reconnect db")
        time.Sleep(5 * time.Second)
	    db.pool, err = pgxpool.ConnectConfig(context.Background(), poolConfig)
    }

	return db, err
}

