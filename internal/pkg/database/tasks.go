package database

import (
	"context"
    "crypto/md5"
    "fmt"
    "io"
	//"github.com/jackc/pgx/v4"
)

func (db *Db) AddURLTask(url string) (string, error) {
	var insertkey string
    h := md5.New()
    io.WriteString(h, url)
    key := fmt.Sprintf("%x", h.Sum(nil))
	err := db.pool.QueryRow(context.Background(), `insert into urls(key, url) values ($1, $2)  RETURNING key;`, key, url).Scan(&insertkey)
	return insertkey, err
}
