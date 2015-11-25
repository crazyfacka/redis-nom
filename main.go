package main

import (
	"os"
	"os/signal"
	"syscall"

	l2n "github.com/crazyfacka/log2nsq"
	"github.com/crazyfacka/redis-nom/dispatcher"
	"github.com/crazyfacka/redis-nom/monitor"

	"code.google.com/p/gcfg"
)

const conffile string = "conf.gcfg"

var cfg config
var masterPodList map[string]*pod

func loadConfiguration() error {
	return gcfg.ReadFileInto(&cfg, conffile)
}

func main() {
	logger2Nsq := l2n.NewLog2Nsq(&l2n.Options{
		AppName: "redis-nom",
	})

	l2n.Println("Starting Redis NOM NOM NOM...")

	if err := loadConfiguration(); err != nil {
		l2n.Printf("%v", err)
		os.Exit(-1)
	}

	l2n.Println("Configuration loaded")
	l2n.Printf("%+v", cfg)

	if err := getSentinelClient(); err != nil {
		l2n.Printf("%v", err)
		os.Exit(-1)
	}

	if masterPodList = retrievePodList(); masterPodList == nil {
		l2n.Println("Failed to retrieve pod list")
		os.Exit(-1)
	}

	l2n.Printf("%+v", masterPodList)

	nsqLock := make(chan bool, 1)

	dispatcher.InitPagerDutyInterface(cfg.PagerDuty.Key)
	dispatcher.StartNSQProducer(cfg.Nsq.Topic, cfg.Nsq.Address, nsqLock)
	monitor.StartSentinelMonitoring(cfg.Redis.Sentinel)
	monitor.StartRedisMonitoring(cfg.Redis.Sentinel)

	lock := make(chan bool, 1)
	sigRcv := make(chan os.Signal, 1)
	signal.Notify(sigRcv, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	// Routine to catch and handle signals
	go func(sigRcv chan os.Signal) {
		for sig := range sigRcv {
			l2n.Printf("Signal caught: %s", sig.String())
			// TODO Clean all
			nsqLock <- true
			lock <- true
		}
	}(sigRcv)

	<-lock

	logger2Nsq.Close()
}
