package main

import (
	"flag"
    "fmt"
    "log"
	"path/filepath"
	"github.com/spf13/viper"
	"git.a.jhuo.ca/huoju/traitement/internal/pkg/rabbitmq"
)


var (
	pgURL string
    amqpURL string
    queueName string
    baseRetryDelay int
    maxRetries int
)

func loadconf() {
	viper.AddConfigPath(filepath.Dir("./configs/"))
	viper.AddConfigPath(filepath.Dir("."))
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.ReadInConfig()
	pgURL = viper.GetString("PG_URL")
	amqpURL = viper.GetString("AMQP_URL")
	queueName = viper.GetString("QUEUE_NAME")
	baseRetryDelay = viper.GetInt("BASE_RETRY_DELAY")
	maxRetries = viper.GetInt("MAX_RETRIES")
}
func main() {
	flag.Parse()
	loadconf()
    amqpQueue, err := rabbitmq.Init(amqpURL, queueName, baseRetryDelay, maxRetries)
    fmt.Println(amqpQueue)
    fmt.Println(err)

    messageChannel, err := amqpQueue.Consume()

    defer amqpQueue.Close()
    fmt.Println(err)
    //handleError(err, "Can't register consumer")
    stopChan := make(chan bool)
    go func(){
        for d := range messageChannel {
            log.Printf("Received a message: %s",d.Body)
            amqpQueue.Retry(&d)
        }
    }()
    <-stopChan
}
