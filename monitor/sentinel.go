package monitor

import (
	"time"

	l2n "github.com/crazyfacka/log2nsq"
	"github.com/crazyfacka/redis-nom/dispatcher"

	"gopkg.in/redis.v3"
)

func bananas(addr string) {
	var err error
	var psub *redis.PubSub

	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	if err = client.Ping().Err(); err != nil {
		l2n.Printf("Error talking with sentinel @ %s", addr)
		pd("Error connecting to sentinel", map[string]interface{}{
			"Addr": addr,
		})
		return
	}

	setRedisClientName(client)

	if psub, err = client.PSubscribe("*"); err != nil {
		l2n.Printf("Error subscribing to * from sentinel @ %s", addr)
		pd("Error subscribing to sentinel", map[string]interface{}{
			"Addr":  addr,
			"Topic": "*",
		})
		return
	}

	l2n.Printf("Subscribed to * from %s", addr)

	defer psub.Close()

	for {
		if msg, err := psub.ReceiveMessage(); err == nil {
			dispatcher.Publish("sentinel", addr, msg.Channel, msg.Payload)
		} else {
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// StartSentinelMonitoring starts the monitoring of sentinels
func StartSentinelMonitoring(addrs []string) {
	for _, addr := range addrs {
		go bananas(addr)
	}
}

func init() {}
