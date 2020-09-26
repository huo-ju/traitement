package task

import (
    "fmt"
    "log"
    "encoding/json"
    "github.com/google/uuid"
    "git.a.jhuo.ca/huoju/traitement/internal/pkg/database"
	"git.a.jhuo.ca/huoju/traitement/pkg/rabbitmq"
    "git.a.jhuo.ca/huoju/traitement/pkg/types"
)

func AddURLMetaTasks(urlmetalist []types.UrlMeta, amqpQueue *rabbitmq.Queue) (string, error) {
    for _, urlmeta := range urlmetalist{
        fmt.Println(urlmeta.Url)
        key,err := database.DBConn.AddURLTask(urlmeta.Url)
        if err  == nil {
            log.Printf("AddURLMetaTask: %s %s", key, urlmeta.Url)
            fmt.Println("===json marshal")
            //metastr := fmt.Sprintf("{\"url\":\"%s\"}", urlmeta.Url)
            buff, err := json.Marshal(urlmeta)
            if err == nil {
                metastr := string(buff)
                fmt.Println(metastr)
                atask := &types.Task{ID: uuid.New().String(), Type:"SPIDER", Meta: metastr}
	            body, err := json.Marshal(atask)
                if err == nil {
                    amqpQueue.Publish(body)
                }
            }else {
                fmt.Println("json format error", err)
            }
        }
        //h := md5.New()
        //io.WriteString(h, urlmeta.Url)
        //key := fmt.Sprintf("%x", h.Sum(nil))
    }
	//var insertkey string
    //h := md5.New()
    //io.WriteString(h, url)
    //key := fmt.Sprintf("%x", h.Sum(nil))
	//err := db.pool.QueryRow(context.Background(), `insert into urls(key, url) values ($1, $2)  RETURNING key;`, key, url).Scan(&insertkey)
	//return insertkey, err
    return "",nil
}
