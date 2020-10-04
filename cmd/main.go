package main

import (
	"flag"
    "fmt"
    "time"
    "github.com/labstack/gommon/log"
	"path/filepath"
	"github.com/spf13/viper"
	"github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "github.com/streadway/amqp"
    rcrabbitmq "github.com/virushuo/go-amqp-reconnect/rabbitmq"
	"git.a.jhuo.ca/huoju/traitement/pkg/rabbitmq"
	"git.a.jhuo.ca/huoju/traitement/api"
	"git.a.jhuo.ca/huoju/traitement/internal/pkg/database"
	//"git.a.jhuo.ca/huoju/traitement/pkg/types"
)


var (
	pgURL string
    amqpURL string
    queueName string
    baseRetryDelay int
    maxRetries int
    jwtSecret   string
	denyDomains map[string]int
)

var amqpQueue *rabbitmq.Queue

func loadconf() {
	viper.AddConfigPath(filepath.Dir("./configs/"))
	viper.AddConfigPath(filepath.Dir("."))
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error parsing config file: %s \n", err))
	}

	pgURL = viper.GetString("PG_URL")
	amqpURL = viper.GetString("AMQP_URL")
	jwtSecret = viper.GetString("JWT_SECRET")
	queueName = viper.GetString("QUEUE_NAME")
	baseRetryDelay = viper.GetInt("BASE_RETRY_DELAY")
	maxRetries = viper.GetInt("MAX_RETRIES")

	viper.SetConfigName("domain")
	viper.SetConfigType("yaml")
	err = viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error parsing config file: %s \n", err))
	}
    denyDomains = make(map[string]int)

    domains := viper.GetStringSlice("deny")
    for _,domain := range domains{
        denyDomains[domain]=1
    }
}

func StartServer(jwtSecret string ) {
    api.DenyDomains = denyDomains
	e := echo.New()
    e.Use(middleware.Logger())
    e.Logger.SetLevel(log.DEBUG)
    r := e.Group("/api")
	r.Use(middleware.JWT([]byte(jwtSecret)))
    r.POST("/addurl", api.AddUrl(amqpQueue))
	e.Logger.Fatal(e.Start(":1323"))
}

func amqpQueueConnect(connectstr string, name string, baseRetryDelay int, maxRetries int, chAmqpErr chan *amqp.Error) (*rabbitmq.Queue){

    amqpconfig := &rcrabbitmq.Config{}
    amqpQueue, err := rabbitmq.Init(amqpURL, queueName, baseRetryDelay, maxRetries, amqpconfig, chAmqpErr)
    for err != nil {
        fmt.Println(err)
        fmt.Println("wait 5 Second for reconnect amqp")
        time.Sleep(5 * time.Second)
        amqpQueue, err = rabbitmq.Init(amqpURL, queueName, baseRetryDelay, maxRetries, amqpconfig, chAmqpErr)
    }
    fmt.Println("amqp connected")
    return amqpQueue
}

func readAmqpErrorChannel(c chan *amqp.Error, amqpQueue *rabbitmq.Queue) {
    input := <-c
    fmt.Println("readAmqpErrorChannel closing:%s", input)
}

func main() {
	flag.Parse()
	loadconf()
    var err error
	database.DBConn, err = database.New(pgURL)
    if err!=nil {
        fmt.Println("dbconn err:")
        fmt.Println(err)
    }

	var chAmqpErr chan *amqp.Error = make(chan *amqp.Error)
    amqpQueue = amqpQueueConnect(amqpURL, queueName, baseRetryDelay, maxRetries, chAmqpErr)
	go readAmqpErrorChannel(chAmqpErr, amqpQueue)
    defer amqpQueue.Close()
    StartServer(jwtSecret)
}
