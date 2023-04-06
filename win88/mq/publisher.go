package mq

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"games.yol.com/win88/common"
	"github.com/idealeak/goserver/core/broker"
	"github.com/idealeak/goserver/core/broker/rabbitmq"
	"github.com/idealeak/goserver/core/logger"
)

const (
	BACKUP_PATH = "backup"
	TIME_FORMAT = "20060102150405"
)

var ERR_CLOSED = errors.New("publisher is closed")

type item struct {
	topic string
	msg   interface{}
}

type RabbitMQPublisher struct {
	b        broker.Broker
	exchange rabbitmq.Exchange
	url      string
	que      chan *item
	closed   bool
	waitor   sync.WaitGroup
}

func NewRabbitMQPublisher(url string, exchange rabbitmq.Exchange, backlog int) *RabbitMQPublisher {
	if backlog <= 0 {
		backlog = 1
	}
	mq := &RabbitMQPublisher{
		url:      url,
		exchange: exchange,
		que:      make(chan *item, backlog),
	}

	rabbitmq.DefaultRabbitURL = mq.url
	rabbitmq.DefaultExchange = mq.exchange

	mq.b = rabbitmq.NewBroker()
	mq.b.Init()
	return mq
}

func (p *RabbitMQPublisher) Start() (err error) {
	if ok, _ := common.PathExists(BACKUP_PATH); !ok {
		err = os.MkdirAll(BACKUP_PATH, os.ModePerm)
		if err != nil {
			return
		}
	}

	err = p.b.Connect()
	if err != nil {
		return
	}

	go p.workerRoutine()

	return nil
}

func (p *RabbitMQPublisher) Stop() error {
	if p.closed {
		return ERR_CLOSED
	}

	p.closed = true
	close(p.que)
	for item := range p.que {
		p.publish(item.topic, item.msg)
	}

	//等待所有投递出去的任务全部完成
	p.waitor.Wait()

	return p.b.Disconnect()
}

func (p *RabbitMQPublisher) Send(topic string, msg interface{}) (err error) {
	if p.closed {
		return ERR_CLOSED
	}

	i := &item{topic: topic, msg: msg}
	select {
	case p.que <- i:
	default:
		//会不会情况更糟糕
		go p.concurrentPublish(topic, msg)
	}
	return nil
}

func (p *RabbitMQPublisher) concurrentPublish(topic string, msg interface{}) (err error) {
	p.waitor.Add(1)
	defer p.waitor.Done()
	return p.publish(topic, msg)
}

func (p *RabbitMQPublisher) publish(topic string, msg interface{}) (err error) {
	defer func() {
		if err != nil {
			p.backup(topic, msg, err)
		}

		recover()
	}()

	buf, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	err = p.b.Publish(topic, &broker.Message{Body: buf})
	if err != nil {
		logger.Logger.Error("RabbitMQPublisher.publish err:", err)
		return
	}
	return nil
}

func (p *RabbitMQPublisher) workerRoutine() {
	p.waitor.Add(1)
	defer p.waitor.Done()

	for {
		select {
		case item, ok := <-p.que:
			if ok {
				p.publish(item.topic, item.msg)
			} else {
				return
			}
		}
	}
}

func (p *RabbitMQPublisher) backup(topic string, msg interface{}, err error) {
	buf, err := json.Marshal(msg)
	if err != nil {
		return
	}
	tNow := time.Now()
	filePath := fmt.Sprintf("%s/%s_%s_%09d_%04d.dat", BACKUP_PATH, topic, tNow.Format(TIME_FORMAT), tNow.Nanosecond(), rand.Int31n(10000))
	f, err := os.Create(filePath)
	if err != nil {
		logger.Logger.Errorf("RabbitMQPublisher.public(%s,%v) err:%v", topic, msg, err)
		return
	}
	defer f.Close()
	var reason string
	if err != nil {
		reason = err.Error()
	}
	f.WriteString("reason:" + reason + "\n")
	f.WriteString("data:" + string(buf) + "\n")
}
