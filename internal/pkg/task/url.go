package task

import (
    "fmt"
    "log"
    "encoding/json"
    "net/url"
    "github.com/google/uuid"
    "git.a.jhuo.ca/huoju/traitement/internal/pkg/database"
	"git.a.jhuo.ca/huoju/traitement/pkg/rabbitmq"
    "git.a.jhuo.ca/huoju/traitement/pkg/types"
)

func AddURLMetaTasks(urlmetalist []types.UrlMeta, denydomains *map[string]int, amqpQueue *rabbitmq.Queue) (string, error) {
    for _, urlmeta := range urlmetalist{
        fmt.Println("addurl: ",urlmeta.Url)
        u, err := url.Parse(urlmeta.Url)
        if err !=nil{
            fmt.Println("Parse url error, skip.", urlmeta.Url)
            continue
        }

        if (*denydomains)[u.Host]==1{
            fmt.Println("domain in the denylist, skip.", urlmeta.Url)
            continue
        }
        if urlmeta.Uniq == true {
            key, err := database.DBConn.AddURLTask(urlmeta.Url)
            if err != nil {
                fmt.Println("insert url to db error")
                fmt.Println(err)
                continue
            }
            log.Printf("AddURLMetaTask To Db: %s %s", key, urlmeta.Url)
        }
        buff, err := json.Marshal(urlmeta)
        if err == nil {
            metastr := string(buff)
            atask := &types.Task{ID: uuid.New().String(), Type:"SPIDER", Meta: metastr}
	        body, err := json.Marshal(atask)
            if err == nil {
                log.Printf("Add URL Task To queue:", urlmeta.Url)
                err := amqpQueue.Publish(body)
                if err!=nil{
                    fmt.Println("publish err:")
                    fmt.Println(err)
                }
            }
        }else {
            fmt.Println("json format error", err)
        }
    }
    return "",nil
}
