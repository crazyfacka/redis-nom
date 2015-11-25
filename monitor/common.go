package monitor

import (
	"gopkg.in/redis.v3"
	"github.com/crazyfacka/redis-nom/dispatcher"
)

type pod struct {
	Addr       string
	Slaves     []string
	SlaveCount int
}

var pd = dispatcher.SendPush

func setRedisClientName(client *redis.Client) *redis.StringCmd {
	cmd := redis.NewStringCmd("CLIENT", "SETNAME", "redis-nom")
	client.Process(cmd)
	return cmd
}
