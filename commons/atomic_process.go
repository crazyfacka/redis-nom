package commons

import (
	"os"
	"time"

	"gopkg.in/redis.v3"

	l2n "github.com/crazyfacka/log2nsq"
)

var cache *redis.Client

func logImAlive() {
	for {
		time.Sleep(time.Minute)
		l2n.Println("[HOLD] STOP! Don't kill me, I'm listening quietly...")
	}
}

func iWillSurvive(env string) {
	l2n.Println("[HOLD] First I was afraid, then I was petrified...")
	for {
		cache.Set(env, "I Will Survive!", 4*time.Second)
		time.Sleep(3 * time.Second)
	}
}

// HoldMeTillICantGetEnough will hold the application here, until he finds out he was left out alone to do all the work
func HoldMeTillICantGetEnough(mastername string, sentinel []string, env string) {
	cache = redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    mastername,
		SentinelAddrs: sentinel,
	})

	if _, err := cache.Ping().Result(); err != nil {
		l2n.Printf("[CACHE] %s", err.Error())
		os.Exit(1)
	}

	monkeyOutThere := false
	for cache.Get(env).Err() == nil {
		if !monkeyOutThere {
			l2n.Println("[HOLD] There's a monkey doing monkey's work...")
			monkeyOutThere = true
		}
		time.Sleep(5 * time.Second)
	}

	go iWillSurvive(env)
	go logImAlive()
}

func init() {}
