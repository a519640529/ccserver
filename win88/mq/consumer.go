package mq

import (
	"fmt"
	"games.yol.com/win88/common"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/idealeak/goserver/core/broker"
	"github.com/idealeak/goserver/core/broker/rabbitmq"
	"github.com/idealeak/goserver/core/logger"
)

var subscriberLock sync.RWMutex
var subscriber = make(map[string][]*Subscriber)

type Subscriber struct {
	broker.Subscriber
	h    broker.Handler
	opts []broker.SubscribeOption
}

func RegisteSubscriber(topic string, h broker.Handler, opts ...broker.SubscribeOption) {
	s := Subscriber{
		h:    h,
		opts: opts,
	}
	subscriberLock.Lock()
	subscriber[topic] = append(subscriber[topic], &s)
	subscriberLock.Unlock()
}

func UnregisteSubscriber(topic string) {
	subscriberLock.Lock()
	delete(subscriber, topic)
	subscriberLock.Unlock()
}

func GetSubscriber(topic string) []*Subscriber {
	subscriberLock.RLock()
	defer subscriberLock.RUnlock()
	if s, ok := subscriber[topic]; ok {
		return s
	}
	return nil
}

func GetSubscribers() map[string][]*Subscriber {
	ret := make(map[string][]*Subscriber)
	subscriberLock.RLock()
	defer subscriberLock.RUnlock()
	for topic, s := range subscriber {
		temp := make([]*Subscriber, len(s))
		copy(temp, s)
		ret[topic] = temp
	}
	return ret
}

type RabbitMQConsumer struct {
	broker.Broker
	url      string
	exchange rabbitmq.Exchange
}

func NewRabbitMQConsumer(url string, exchange rabbitmq.Exchange) *RabbitMQConsumer {
	mq := &RabbitMQConsumer{
		url:      url,
		exchange: exchange,
	}

	rabbitmq.DefaultRabbitURL = mq.url
	rabbitmq.DefaultExchange = mq.exchange

	mq.Broker = rabbitmq.NewBroker()
	mq.Broker.Init()
	return mq
}

func (c *RabbitMQConsumer) Start() error {
	if err := c.Connect(); err != nil {
		return err
	}

	sss := GetSubscribers()
	for topic, ss := range sss {
		for _, s := range ss {
			sub, err := c.Subscribe(topic, s.h, s.opts...)
			if err != nil {
				return err
			}

			s.Subscriber = sub
		}
	}

	return nil
}

func (c *RabbitMQConsumer) Stop() error {
	sss := GetSubscribers()
	for _, ss := range sss {
		for _, s := range ss {
			s.Unsubscribe()
		}
	}
	return c.Disconnect()
}

func BackUp(e broker.Event, err error) {
	tNow := time.Now()
	filePath := fmt.Sprintf("%s/%s_%s_%09d_%04d.dat", BACKUP_PATH, e.Topic(), tNow.Format(TIME_FORMAT), tNow.Nanosecond(), rand.Int31n(10000))
	f, err := os.Create(filePath)
	if err != nil {
		logger.Logger.Errorf("RabbitMQPublisher.public(%s,%v) err:%v", e.Topic(), e.Message(), err)
		return
	}
	defer f.Close()
	var reason string
	if err != nil {
		reason = err.Error()
	}
	f.WriteString("reason:" + reason + "\n")
	f.WriteString("data:" + string(e.Message().Body) + "\n")
}

func init() {
	if ok, _ := common.PathExists(BACKUP_PATH); !ok {
		os.MkdirAll(BACKUP_PATH, os.ModePerm)
	}
}
