package main

import (
	"flag"
    "fmt"
    "log"
	"path/filepath"
    "encoding/json"
	"github.com/spf13/viper"
	"git.a.jhuo.ca/huoju/traitement/pkg/rabbitmq"
	"git.a.jhuo.ca/huoju/traitement/pkg/types"
)


var (
	pgURL string
    amqpURL string
    //queueName string
    queueQos int
    baseRetryDelay int
    maxRetries int
)

type SpiderTask struct {
    Url string
}


func loadconf() {
	viper.AddConfigPath(filepath.Dir("./configs/"))
	viper.AddConfigPath(filepath.Dir("."))
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.ReadInConfig()
	pgURL = viper.GetString("PG_URL")
	amqpURL = viper.GetString("AMQP_URL")
	//queueName = viper.GetString("QUEUE_NAME")
	baseRetryDelay = viper.GetInt("BASE_RETRY_DELAY")
	maxRetries = viper.GetInt("MAX_RETRIES")
    queueQos = viper.GetInt("QUEUE_QOS")
}

func main() {
	flag.Parse()
	loadconf()
    queueName := "spider"
    amqpQueue, err := rabbitmq.Init(amqpURL, queueName, baseRetryDelay, maxRetries)
    messageChannel, err := amqpQueue.Consume(2)
    defer amqpQueue.Close()
    fmt.Println(err)
    //handleError(err, "Can't register consumer")
    stopChan := make(chan bool)
    go func(){
        for d := range messageChannel {
            atask := &types.Task{}
            err := json.Unmarshal(d.Body, atask)
            log.Println(err)
            log.Printf("Received a message: %s", atask)
            if atask.Type == "SPIDER"  {
                ataskmeta := &SpiderTask{}
                err = json.Unmarshal([]byte(atask.Meta), ataskmeta)
                log.Printf("Message Meta: %s", ataskmeta )
            }
            amqpQueue.Retry(&d)
        }
    }()

    <-stopChan
}

