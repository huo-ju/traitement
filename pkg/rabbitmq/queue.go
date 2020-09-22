package rabbitmq

import (
    "fmt"
    //"os"
	"strconv"
	//"flag"
    //"encoding/json"
    //"math/rand"
    //"time"
    //"log"
    "github.com/streadway/amqp"
)

// Queue wrapping the amqp Channel operation and manage the connection.
type Queue struct {
    AmqpChannel *amqp.Channel
    Conn *amqp.Connection
    Name string
    BaseRetryDelay int
    MaxRetries int
}

// Close the channel and connection
func (q *Queue) Close() {
	q.AmqpChannel.Close()
    q.Conn.Close()
}

// Consume wrapping the mailman.created queue consume 
func (q *Queue) Consume() (<-chan amqp.Delivery, error) {
    return q.AmqpChannel.Consume("mailman."+q.Name+".created", "",false, false, false,false, nil)
}

//Retry the job task
func  (q *Queue) Retry(d *amqp.Delivery) {
    QueueName := "mailman."+q.Name+".created"

    if d.Exchange == "nanit."+q.Name {
        QueueName = "mailman."+q.Name+".created"
    } else if d.Exchange == "nanit."+q.Name+".retry1" {
        QueueName = "nanit."+q.Name+".wait_queue"
    } else if d.Exchange == "nanit."+q.Name+".retry2" {
        QueueName = "mailman."+q.Name+".created"
    }

    retryCount := 0
    if d.Headers["x-retries"] != nil {
        xretries, err := strconv.Atoi(fmt.Sprintf("%v", d.Headers["x-retries"]) )
        if err == nil {
            retryCount = xretries
        }
    }

    if retryCount < q.MaxRetries  { //retrycount 3
        retryDelay := q.BaseRetryDelay * (retryCount + 1)
        err := q.AmqpChannel.Publish("nanit."+q.Name+".retry1", QueueName, false, false, amqp.Publishing{
		    DeliveryMode: amqp.Persistent,
            Expiration: strconv.Itoa(retryDelay),
            Headers: amqp.Table{"x-retries" : retryCount + 1},
		    ContentType:  "text/plain",
		    Body:         d.Body,
		})
        if err ==nil {
            d.Ack(false)
        }
    }else {
        //no more retry
        //ack the task
        err := q.AmqpChannel.Publish("", "mailman."+q.Name+".rejected", false, false, amqp.Publishing{
		    DeliveryMode: amqp.Persistent,
            Headers: amqp.Table{"x-retries" : retryCount + 1},
		    ContentType:  "text/plain",
		    Body:         d.Body,
		})
        if err ==nil {
            d.Ack(false)
        }

    }
}

// Publish send task to the queue
func  (q *Queue) Publish(body []byte) error {
	return q.AmqpChannel.Publish("nanit."+q.Name, "created", false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "text/plain",
		Body:         body,
	})
}

// Init the Queue and return a Queue instance
func Init (connectstr string, name string, baseRetryDelay int, maxRetries int) (*Queue, error)  {
    conn, err := amqp.Dial(connectstr)
    if err !=nil {
        return nil, err
    }
	//defer conn.Close()

	amqpChannel, err := conn.Channel()
    if err !=nil {
        return nil, err
    }
	//defer amqpChannel.Close()

    err = amqpChannel.ExchangeDeclare("nanit."+name , "direct", true, false, false, false, nil); //"nanit.users"
    if err !=nil {
        return nil, err
    }
	_, err = amqpChannel.QueueDeclare("mailman."+name+".created", true, false, false, false, nil)
    if err !=nil {
        return nil, err
    }
	_, err = amqpChannel.QueueDeclare("mailman."+name+".rejected", true, false, false, false, nil)
    if err !=nil {
        return nil, err
    }
    err = amqpChannel.ExchangeDeclare("nanit."+name+".retry1", "direct", true, false, false, false, nil);
    if err !=nil {
        return nil, err
    }
    err = amqpChannel.ExchangeDeclare("nanit."+name+".retry2", "direct", true, false, false, false, nil);
    if err !=nil {
        return nil, err
    }
    waitqargs := make(amqp.Table)
    waitqargs["x-dead-letter-exchange"] = "nanit."+name+".retry2"
	_, err = amqpChannel.QueueDeclare("nanit."+name+".wait_queue", true, false, false, false, waitqargs)
    if err !=nil {
        return nil, err
    }
    err = amqpChannel.QueueBind("mailman."+name+".created", "created","nanit."+name,false , nil);
    if err !=nil {
        return nil, err
    }
    err = amqpChannel.QueueBind("nanit."+name+".wait_queue", "mailman."+name+".created", "nanit."+name+".retry1",false , nil);
    if err !=nil {
        return nil, err
    }
    err = amqpChannel.QueueBind("mailman."+name+".created", "mailman."+name+".created", "nanit."+name+".retry2",false , nil);
    if err !=nil {
        return nil, err
    }
    queue := &Queue{AmqpChannel: amqpChannel, Conn:conn, Name: name, BaseRetryDelay: baseRetryDelay, MaxRetries: maxRetries}
    return queue, nil
}
