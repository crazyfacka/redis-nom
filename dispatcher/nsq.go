package dispatcher

import (
	"encoding/json"
	"fmt"
	"time"

	l2n "github.com/crazyfacka/log2nsq"

	"github.com/bitly/go-nsq"
)

type log struct {
	Who     string      `json:"who"`
	Where   string      `json:"where"`
	What    string      `json:"what"`
	When    time.Time   `json:"when"`
	Details interface{} `json:"details"`
}

var topic string
var producers map[string]*nsq.Producer

// Publish send this to the interwebs
func Publish(who string, where string, what string, details interface{}) {
	if len(producers) == 0 {
		return
	}

	msg := &log{
		Who:     who,
		Where:   where,
		What:    what,
		When:    time.Now(),
		Details: details,
	}

	for _, producer := range producers {
		if jsonMsg, err := json.Marshal(msg); err == nil {
			if err = producer.Publish(topic, jsonMsg); err != nil {
				l2n.Printf("Error sending message: %s", string(jsonMsg))
			}
		}
	}
}

func bananas(addrs []string, lock chan bool) {
	cfg := nsq.NewConfig()
	cfg.UserAgent = fmt.Sprintf("redis-nom/0.1 go-nsq/%s", nsq.VERSION)

	producers = make(map[string]*nsq.Producer)
	for _, addr := range addrs {
		if producer, err := nsq.NewProducer(addr, cfg); err == nil {
			producers[addr] = producer
		} else {
			l2n.Printf("Failed to connect to NSQ @ %s", addr)
		}
	}

	if len(producers) == 0 {
		l2n.Println("No NSQ available")
		return
	}

	l2n.Println("NSQ Producer is on the go!")

	<-lock

	for _, producer := range producers {
		producer.Stop()
	}
}

// StartNSQProducer go go NSQ
func StartNSQProducer(t string, addrs []string, lock chan bool) {
	topic = t
	go bananas(addrs, lock)
}

func init() {}
