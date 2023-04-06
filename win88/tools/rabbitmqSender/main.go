package main

import (
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"hello", // name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	isok := true
	c := make(chan *amqp.Error, 1)
	ch.NotifyClose(c)
	go func() {
		for {
			select {
			case e, ok := <-c:
				fmt.Println("NotifyClose", ok, e)
				isok = false
				return
			}
		}
	}()
	i := 0
	for isok {
		i++
		body := fmt.Sprintf("hello_lyk%v", i)
		err = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType:  "text/plain",
				Body:         []byte(body),
				DeliveryMode: amqp.Persistent,
			})
		log.Printf(" [x] Sent %s", body)
		//failOnError(err, "Failed to publish a message")
		if i == 1000 {
			break
		}
		time.Sleep(time.Second)
	}

}
