package main

import (
	"flag"
    "fmt"
    "log"
	"path/filepath"
    "encoding/json"
	"github.com/spf13/viper"
	"github.com/gocolly/colly/v2"
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
	GatherLink bool
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
                log.Printf("Message Url: %s", ataskmeta.Url )

				c := colly.NewCollector(
					colly.UserAgent("Mozilla/5.0 (compatible; traitementBot; http://opentraitement.org)"),
				)
				c.OnHTML("a[href]", func(e *colly.HTMLElement) {
                    if ataskmeta.GatherLink == true{ //TOOD: gather links and sent to the api
						link := e.Attr("href")
						// Print link
						fmt.Printf("Link found: %q -> %s\n", e.Text, link)
					}
				})

				c.OnResponse(func(r *colly.Response) {
					fmt.Println(string(r.Body))
					fmt.Println("response code: ", r.StatusCode)
                    //TODO: 200 extract and save file, send ack to the queue
				})

				// Before making a request print "Visiting ..."
				c.OnRequest(func(r *colly.Request) {
					fmt.Println("Visiting", r.URL.String())
				})

				// Start scraping on https://hackerspaces.org
				c.Visit(ataskmeta.Url)
            }
            //amqpQueue.Succ(&d)
            //amqpQueue.Retry(&d)
        }
    }()

    <-stopChan
}

