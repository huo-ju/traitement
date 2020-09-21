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
}
func main() {
	flag.Parse()
	loadconf()
    fmt.Println(pgURL)
    fmt.Println(amqpURL)

    //msg := new(briko.Message)

    amqpQueue, err := rabbitmq.Init(amqpURL, queueName)
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
        }
    }()
    <-stopChan
}
