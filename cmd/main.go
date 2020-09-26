package main

import (
	"flag"
    "fmt"
    "github.com/labstack/gommon/log"
	"path/filepath"
	"github.com/spf13/viper"
	"github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
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

func main() {
	flag.Parse()
	loadconf()
    fmt.Println(database.DBConn)
    var err error
	database.DBConn, err = database.New(pgURL)
    fmt.Println("dbconn:")
    fmt.Println(database.DBConn)

    //r,err := db.AddURLTask("https://google.com")
    //fmt.Println(r)
    //fmt.Println(err)

    amqpQueue, err = rabbitmq.Init(amqpURL, queueName, baseRetryDelay, maxRetries)
    //messageChannel, err := amqpQueue.Consume(10) //TODO: Qos count add to config file
    defer amqpQueue.Close()
    fmt.Println(err)
    //handleError(err, "Can't register consumer")
    //stopChan := make(chan bool)
    //go func(){
    //    for d := range messageChannel {
    //        log.Printf("Received a message: %s",d.Body)
    //        amqpQueue.Retry(&d)
    //    }
    //}()

    //amqpQueueStorage, err := rabbitmq.Init(amqpURL, "storage", baseRetryDelay, maxRetries)
    //messageChannelStorage, err := amqpQueueStorage.Consume()
    //defer amqpQueueStorage.Close()

    //go func(){
    //    for d := range messageChannelStorage{
    //        log.Printf("Received a storage message: %s",d.Body)
    //        amqpQueueStorage.Retry(&d)
    //    }
    //}()

    //<-stopChan

    StartServer(jwtSecret)
}
