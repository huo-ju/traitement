package main

import (
	"flag"
    "fmt"
    "log"
    "time"
	"path/filepath"
    "encoding/json"
	"github.com/spf13/viper"
	"github.com/gocolly/colly/v2"
	"git.a.jhuo.ca/huoju/traitement/pkg/rabbitmq"
	"git.a.jhuo.ca/huoju/traitement/pkg/types"
    "git.a.jhuo.ca/huoju/traitement/internal/pkg/html"
	"git.a.jhuo.ca/huoju/traitement/pkg/storage"
)


var (
	pgURL string
    amqpURL string
    //queueName string
    queueQos int
    baseRetryDelay int
    maxRetries int
    webapiendpoint string
    webtoken string
    sleeptime int
)

type SpiderTask struct {
    Url string
	GatherLink bool
    Uniq bool
    SavePage bool
}


func loadconf() {
	viper.AddConfigPath(filepath.Dir("./configs/"))
	viper.AddConfigPath(filepath.Dir("."))
	viper.SetConfigName("worker_config")
	viper.SetConfigType("toml")
	viper.ReadInConfig()
	pgURL = viper.GetString("PG_URL")
	amqpURL = viper.GetString("AMQP_URL")
	//queueName = viper.GetString("QUEUE_NAME")
	baseRetryDelay = viper.GetInt("BASE_RETRY_DELAY")
	maxRetries = viper.GetInt("MAX_RETRIES")
    queueQos = viper.GetInt("QUEUE_QOS")
    webapiendpoint = viper.GetString("WEBAPI")
    webtoken = viper.GetString("JWT_TOKEN")
    sleeptime= viper.GetInt("SLEEP")
}

func main() {
	flag.Parse()
	loadconf()
    fmt.Println(amqpURL)
    amqpQueue, err := rabbitmq.Init(amqpURL, queueName, baseRetryDelay, maxRetries)
    fmt.Println(err)
    messageChannel, err := amqpQueue.Consume(2)
    defer amqpQueue.Close()
    fmt.Println(err)
    //handleError(err, "Can't register consumer")
    stopChan := make(chan bool)
    go func(){
        for d := range messageChannel {
            fmt.Println("sleep...")
            time.Sleep(time.Duration(sleeptime) * time.Millisecond)
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
				//c.OnHTML("a[href]", func(e *colly.HTMLElement) {
                //    if ataskmeta.GatherLink == true{ //TOOD: gather links and sent to the api
				//		link := e.Attr("href")
				//		// Print link
				//		fmt.Printf("Link found: %q -> %s\n", e.Text, link)
				//	}
				//})
                c.OnError(func(r *colly.Response, err error) {
		            fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
                    amqpQueue.Retry(&d)
	            })

				c.OnResponse(func(r *colly.Response) {
					fmt.Println("response code: ", r.StatusCode)
                    if r.StatusCode == 200 {
                        succ := true
                        if ataskmeta.SavePage == true {
                            bucket := &storage.FileBucket{"/home/huoju/crawling"}

                            pagecontent,err := html.FindContent(r.Request.URL.String(), string(r.Body))
                            fileinfo, err := bucket.Save(pagecontent, atask.ID)
                            if err != nil{
                                fmt.Println(err)
                                succ = false
                            }else{
                                fmt.Println("save succ", fileinfo)
                            }
                        }
                        if ataskmeta.GatherLink == true {
                            urlprefix := fmt.Sprintf("%s://%s", r.Request.URL.Scheme,r.Request.URL.Host)
                            urlmeta_list := html.FindLink(urlprefix, string(r.Body))
                            webapi := &storage.WebApi{webapiendpoint,webtoken}
                            err := webapi.SaveUrls(&urlmeta_list)
                            if err != nil {
                                fmt.Println(err)
                                succ = false
                            }else {
                                fmt.Println("submit urls succ")
                            }
                        }

                        if succ == false {
                            //retry
                        } else {
                            amqpQueue.Succ(&d)
                        }

                    }else {
                        amqpQueue.Retry(&d)
                    }
				})

				c.OnRequest(func(r *colly.Request) {
					fmt.Println("Visiting", r.URL.String())
				})

				// Start scraping on https://hackerspaces.org
				c.Visit(ataskmeta.Url)
            }
            //amqpQueue.Succ(&d)
        }
    }()

    <-stopChan
}

