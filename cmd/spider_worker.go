package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"time"
	//"github.com/gocolly/colly/v2"
	"git.a.jhuo.ca/huoju/traitement/internal/pkg/html"
	"git.a.jhuo.ca/huoju/traitement/internal/pkg/scraping"
	"git.a.jhuo.ca/huoju/traitement/pkg/rabbitmq"
	"git.a.jhuo.ca/huoju/traitement/pkg/storage"
	"git.a.jhuo.ca/huoju/traitement/pkg/types"
	"github.com/streadway/amqp"
	rcrabbitmq "github.com/virushuo/go-amqp-reconnect/rabbitmq"
)

var (
	pgURL            string
	amqpURL          string
	queueName        string
	queueQos         int
	baseRetryDelay   int
	maxRetries       int
	webapiendpoint   string
	webtoken         string
	sleeptime        int
	fileStoragePath  string
	chromeRpc        string
	chromeRpcTimeout int
)

var QueueMessageChannel <-chan amqp.Delivery

type SpiderTask struct {
	Url        string
	GatherLink bool
	Uniq       bool
	SavePage   bool
}

func loadconf() {
	fmt.Println("Reading Environment Variable")
	var configspath string
	configspath = os.Getenv("ConfigsPath")
	if configspath == "" {
		configspath = filepath.Dir("./configs/")
	}
	fmt.Printf("configspath: %s\n", configspath)

	//viper.AddConfigPath(filepath.Dir(configspath))
	viper.AddConfigPath(configspath)
	viper.SetConfigName("worker_config")
	viper.SetConfigType("toml")
	err := viper.ReadInConfig()
	fmt.Println("read in config")
	fmt.Println(err)
	pgURL = viper.GetString("PG_URL")
	amqpURL = viper.GetString("AMQP_URL")
	fileStoragePath = viper.GetString("FILE_STORAGE")
	queueName = viper.GetString("QUEUE_NAME")
	baseRetryDelay = viper.GetInt("BASE_RETRY_DELAY")
	maxRetries = viper.GetInt("MAX_RETRIES")
	queueQos = viper.GetInt("QUEUE_QOS")
	webapiendpoint = viper.GetString("WEBAPI")
	webtoken = viper.GetString("JWT_TOKEN")
	sleeptime = viper.GetInt("SLEEP")
	chromeRpc = viper.GetString("CHROME_DEBUG_RPC")
	chromeRpcTimeout = viper.GetInt("CHROME_DEBUG_RPC_TIMEOUT")
}

func amqpQueueConnect(connectstr string, name string, baseRetryDelay int, maxRetries int, chAmqpErr chan *amqp.Error) *rabbitmq.Queue {

	amqpconfig := &rcrabbitmq.Config{Qos: queueQos}
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

func processTask(d amqp.Delivery, amqpQueue *rabbitmq.Queue) {
	time.Sleep(time.Duration(sleeptime) * time.Millisecond)
	atask := &types.Task{}
	err := json.Unmarshal(d.Body, atask)
	log.Println(err)
	log.Printf("Received a message: %s", atask)
	if atask.Type == "SPIDER" {
		ataskmeta := &SpiderTask{}
		err = json.Unmarshal([]byte(atask.Meta), ataskmeta)
		log.Printf("Message Meta: %s", ataskmeta)
		log.Printf("Message Url: %s", ataskmeta.Url)
		chromedevt := scraping.New(chromeRpc, time.Duration(chromeRpcTimeout)*time.Second)
		resulthtml, err := chromedevt.Fetch(ataskmeta.Url)
		if err != nil {
			fmt.Println("Request URL:", ataskmeta.Url, "failed with response:", err, "\nError:", err)
			amqpQueue.Retry(&d)
		} else {
			succ := true
			if ataskmeta.SavePage == true {
				bucket := &storage.FileBucket{fileStoragePath}

				pagecontent, err := html.FindContent(ataskmeta.Url, resulthtml)
				fileinfo, err := bucket.Save(pagecontent, atask.ID)
				if err != nil {
					fmt.Println(err)
					succ = false
				} else {
					fmt.Println("save succ", fileinfo)
				}

				pagecontent.Content = resulthtml

				fileinfo, err = bucket.Save(pagecontent, atask.ID+".html")
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println("save html succ", fileinfo)
				}

			}
			if ataskmeta.GatherLink == true {
				u, err := url.Parse(ataskmeta.Url)
				urlprefix := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
				urlmeta_list := html.FindLink(urlprefix, resulthtml)
				webapi := &storage.WebApi{webapiendpoint, webtoken}
				err = webapi.SaveUrls(&urlmeta_list)
				if err != nil {
					fmt.Println(err)
					succ = false
				} else {
					fmt.Println("submit urls succ")
				}
			}

			if succ == false {
				//retry
			} else {
				amqpQueue.Succ(&d)
			}
		}
	}
}

func main() {
	flag.Parse()
	loadconf()

	var chAmqpErr chan *amqp.Error = make(chan *amqp.Error)
	var err error

	fmt.Println("amqpURL: ", amqpURL)
	amqpQueue := amqpQueueConnect(amqpURL, queueName, baseRetryDelay, maxRetries, chAmqpErr)
	QueueMessageChannel, err = amqpQueue.Consume(2)
	if err != nil {
		fmt.Println("can't register Consume err")
	}
	defer amqpQueue.Close()

	go readAmqpErrorChannel(chAmqpErr, amqpQueue)
	go func() {
		for d := range QueueMessageChannel {
			processTask(d, amqpQueue)
		}
		fmt.Println("routine channel closed, exit")
		return
	}()
	stopChan := make(chan bool)
	<-stopChan
}
