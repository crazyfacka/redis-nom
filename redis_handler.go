package main

import (
	"errors"
	"strconv"
	"strings"

	l2n "github.com/crazyfacka/log2nsq"

	"gopkg.in/redis.v3"
)

var client *redis.Client

func setRedisClientName(client *redis.Client) *redis.StringCmd {
	cmd := redis.NewStringCmd("CLIENT", "SETNAME", "redis-nom")
	client.Process(cmd)
	return cmd
}

func retrieveSlaves(master string, pod *pod) {
	Slaves := func(client *redis.Client, master string) *redis.SliceCmd {
		cmd := redis.NewSliceCmd("SENTINEL", "slaves", master)
		client.Process(cmd)
		return cmd
	}

	if rsp, err := Slaves(client, master).Result(); err == nil {
		for _, item := range rsp {
			pod.Slaves = append(pod.Slaves, item.([]interface{})[1].(string))
		}
	}
}

func retrievePodList() map[string]*pod {
	podList := make(map[string]*pod)

	Pods := func(client *redis.Client) *redis.StringCmd {
		cmd := redis.NewStringCmd("INFO", "sentinel")
		client.Process(cmd)
		return cmd
	}

	if rsp, err := Pods(client).Result(); err == nil {
		lines := strings.Split(rsp, "\r\n")
		for i, line := range lines {
			if i > 4 && len(line) > 0 {
				contents1 := strings.SplitN(line, ":", 2)
				contents2 := strings.Split(contents1[1], ",")

				var slaveCount int
				if slaveCount, err = strconv.Atoi(contents2[3][7:]); err != nil {
					slaveCount = -1
				}

				podList[contents2[0][5:]] = &pod{
					Addr:       contents2[2][8:],
					SlaveCount: slaveCount,
				}
			}
		}

	} else {
		l2n.Printf("%v", err)
		return nil
	}

	return podList
}

func getSentinelClient() error {
	foundValidSentinel := false

	for i := 0; i < len(cfg.Redis.Sentinel) && foundValidSentinel == false; i++ {
		client = redis.NewClient(&redis.Options{
			Addr: cfg.Redis.Sentinel[i],
		})

		if err := client.Ping().Err(); err == nil {
			foundValidSentinel = true
		}
	}

	if !foundValidSentinel {
		return errors.New("No valid Sentinel found")
	}

	setRedisClientName(client)

	return nil
}
