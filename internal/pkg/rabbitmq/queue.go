package rabbitmq

import (
    "fmt"
    //"os"
	//"strconv"
	//"flag"
    //"encoding/json"
    //"math/rand"
    //"time"
    //"log"
    "github.com/streadway/amqp"
)


type Queue struct {
    AmqpChannel *amqp.Channel
    Conn *amqp.Connection
    Name string

}


func (q *Queue) Close() {
	q.AmqpChannel.Close()
    q.Conn.Close()
}

func (q *Queue) Consume() (<-chan amqp.Delivery, error) {
    fmt.Println(q.AmqpChannel)
    return q.AmqpChannel.Consume("mailman."+q.Name+".created", "",false, false, false,false, nil)
}

func Init (connectstr string, name string) (*Queue, error)  {
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
    queue := &Queue{AmqpChannel: amqpChannel, Conn:conn, Name: name, }
    return queue, nil
}
