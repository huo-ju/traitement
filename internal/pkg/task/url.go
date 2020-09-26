package task

import (
    "fmt"
    "log"
    "errors"
    "encoding/json"
    "net/url"
    "github.com/google/uuid"
    "git.a.jhuo.ca/huoju/traitement/internal/pkg/database"
	"git.a.jhuo.ca/huoju/traitement/pkg/rabbitmq"
    "git.a.jhuo.ca/huoju/traitement/pkg/types"
)

func AddURLMetaTasks(urlmetalist []types.UrlMeta, denydomains *map[string]int, amqpQueue *rabbitmq.Queue) (string, error) {
    for _, urlmeta := range urlmetalist{
        fmt.Println(urlmeta.Url)
        u, err := url.Parse(urlmeta.Url)
        if err !=nil{
            return "",err
        }

        if (*denydomains)[u.Host]==1{
            return "", errors.New("domain in the denylist, skip.")
        }

        //TODO: add url deny list
        key,err := database.DBConn.AddURLTask(urlmeta.Url)
        if err  == nil {
            log.Printf("AddURLMetaTask: %s %s", key, urlmeta.Url)
            buff, err := json.Marshal(urlmeta)
            if err == nil {
                metastr := string(buff)
                atask := &types.Task{ID: uuid.New().String(), Type:"SPIDER", Meta: metastr}
	            body, err := json.Marshal(atask)
                if err == nil {
                    amqpQueue.Publish(body)
                }
            }else {
                fmt.Println("json format error", err)
            }
        }
    }
    return "",nil
}
