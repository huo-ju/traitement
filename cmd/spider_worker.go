package main

import (
	"flag"
    "fmt"
    "log"
    "time"
    "net/url"
	"path/filepath"
    "encoding/json"
	"github.com/spf13/viper"
	//"github.com/gocolly/colly/v2"
    "github.com/streadway/amqp"
    rcrabbitmq "github.com/virushuo/go-amqp-reconnect/rabbitmq"
	"git.a.jhuo.ca/huoju/traitement/pkg/rabbitmq"
	"git.a.jhuo.ca/huoju/traitement/pkg/types"
    "git.a.jhuo.ca/huoju/traitement/internal/pkg/html"
    "git.a.jhuo.ca/huoju/traitement/internal/pkg/scraping"
	"git.a.jhuo.ca/huoju/traitement/pkg/storage"
)


var (
	pgURL string
    amqpURL string
    queueName string
    queueQos int
    baseRetryDelay int
    maxRetries int
    webapiendpoint string
    webtoken string
    sleeptime int
    fileStoragePath string
    chromeRpc string
    chromeRpcTimeout int
)

var QueueMessageChannel <-chan amqp.Delivery

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
    fileStoragePath = viper.GetString("FILE_STORAGE")
	queueName = viper.GetString("QUEUE_NAME")
	baseRetryDelay = viper.GetInt("BASE_RETRY_DELAY")
	maxRetries = viper.GetInt("MAX_RETRIES")
    queueQos = viper.GetInt("QUEUE_QOS")
    webapiendpoint = viper.GetString("WEBAPI")
    webtoken = viper.GetString("JWT_TOKEN")
    sleeptime= viper.GetInt("SLEEP")
    chromeRpc = viper.GetString("CHROME_DEBUG_RPC")
    chromeRpcTimeout = viper.GetInt("CHROME_DEBUG_RPC_TIMEOUT")
}

func amqpQueueConnect(connectstr string, name string, baseRetryDelay int, maxRetries int, chAmqpErr chan *amqp.Error) (*rabbitmq.Queue){

    amqpconfig := &rcrabbitmq.Config{Qos:queueQos}
    amqpQueue, err := rabbitmq.Init(amqpURL, queueName, baseRetryDelay, maxRetries, amqpconfig, chAmqpErr)
    for err != nil {
        fmt.Println(err)
        fmt.Println("wait 5 Second for reconnect amqp")
        time.Sleep(5 * time.Second)
        amqpQueue, err = rabbitmq.Init(amqpURL, queueName, baseRetryDelay, maxRetries, amqpconfig,chAmqpErr)
    }
    fmt.Println("amqp connected")
    return amqpQueue
}

func readAmqpErrorChannel(c chan *amqp.Error, amqpQueue *rabbitmq.Queue) {
    input := <-c
    fmt.Println("readAmqpErrorChannel closing:%s", input)
}

func processTask(d amqp.Delivery, amqpQueue *rabbitmq.Queue){
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
        chromedevt := scraping.New(chromeRpc, time.Duration(chromeRpcTimeout) * time.Second)
        resulthtml, err := chromedevt.Fetch(ataskmeta.Url)
        if err != nil {
	        fmt.Println("Request URL:", ataskmeta.Url, "failed with response:", err, "\nError:", err)
            amqpQueue.Retry(&d)
        } else {
            succ := true
            if ataskmeta.SavePage == true {
                bucket := &storage.FileBucket{fileStoragePath}

                pagecontent,err := html.FindContent(ataskmeta.Url, resulthtml)
                fileinfo, err := bucket.Save(pagecontent, atask.ID)
                if err != nil{
                    fmt.Println(err)
                    succ = false
                }else{
                    fmt.Println("save succ", fileinfo)
                }
            }
            if ataskmeta.GatherLink == true {
                u, err := url.Parse(ataskmeta.Url)
                urlprefix := fmt.Sprintf("%s://%s", u.Scheme,u.Host)
                urlmeta_list := html.FindLink(urlprefix, resulthtml)
                webapi := &storage.WebApi{webapiendpoint,webtoken}
                err = webapi.SaveUrls(&urlmeta_list)
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
        }


		//c := colly.NewCollector(
		//	colly.UserAgent("Mozilla/5.0 (compatible; traitementBot; http://opentraitement.org)"),
		//)
        //c.OnError(func(r *colly.Response, err error) {
	    //    fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
        //    amqpQueue.Retry(&d)
	    //})

		//c.OnResponse(func(r *colly.Response) {
		//	fmt.Println("response code: ", r.StatusCode)
        //    if r.StatusCode == 200 {
        //        succ := true
        //        if ataskmeta.SavePage == true {
        //            bucket := &storage.FileBucket{fileStoragePath}

        //            pagecontent,err := html.FindContent(r.Request.URL.String(), string(r.Body))
        //            fileinfo, err := bucket.Save(pagecontent, atask.ID)
        //            if err != nil{
        //                fmt.Println(err)
        //                succ = false
        //            }else{
        //                fmt.Println("save succ", fileinfo)
        //            }
        //        }
        //        if ataskmeta.GatherLink == true {
        //            urlprefix := fmt.Sprintf("%s://%s", r.Request.URL.Scheme,r.Request.URL.Host)
        //            urlmeta_list := html.FindLink(urlprefix, string(r.Body))
        //            webapi := &storage.WebApi{webapiendpoint,webtoken}
        //            err := webapi.SaveUrls(&urlmeta_list)
        //            if err != nil {
        //                fmt.Println(err)
        //                succ = false
        //            }else {
        //                fmt.Println("submit urls succ")
        //            }
        //        }

        //        if succ == false {
        //            //retry
        //        } else {
        //            amqpQueue.Succ(&d)
        //        }

        //    }else {
        //        amqpQueue.Retry(&d)
        //    }
		//})

		//c.OnRequest(func(r *colly.Request) {
		//	fmt.Println("Visiting", r.URL.String())
		//})

		// Start scraping on https://hackerspaces.org
		//c.Visit(ataskmeta.Url)
    }
}

func main() {
	flag.Parse()
	loadconf()


	var chAmqpErr chan *amqp.Error = make(chan *amqp.Error)
    var err error

    amqpQueue := amqpQueueConnect(amqpURL, queueName, baseRetryDelay, maxRetries, chAmqpErr)
    QueueMessageChannel, err = amqpQueue.Consume(2)
    if err!= nil{
    fmt.Println("can't register Consume err")
    }
    defer amqpQueue.Close()

	go readAmqpErrorChannel(chAmqpErr, amqpQueue)
    go func(){
        for d := range QueueMessageChannel{
            processTask(d, amqpQueue)
        }
        fmt.Println("routine channel closed, exit")
        return
    }()
    stopChan := make(chan bool)
    <-stopChan
}

