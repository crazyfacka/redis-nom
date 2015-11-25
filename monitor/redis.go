package monitor

import (
	"errors"
	"strconv"
	"strings"
	"time"

	l2n "github.com/crazyfacka/log2nsq"
	"github.com/crazyfacka/redis-nom/dispatcher"
	"gopkg.in/redis.v3"
)

var infoSections = []string{"clients", "memory", "stats", "replication"}
var masterPodList map[string]*pod

func getInfoAsBlob(client *redis.Client) map[string][]string {
	pipeline := client.Pipeline()
	info := make(map[string][]string)

	cmds := []*redis.StringCmd{redis.NewStringCmd("INFO", infoSections[0]),
		redis.NewStringCmd("INFO", infoSections[1]),
		redis.NewStringCmd("INFO", infoSections[2]),
		redis.NewStringCmd("INFO", infoSections[3])}

	for _, cmd := range cmds {
		pipeline.Process(cmd)
	}

	if _, err := pipeline.Exec(); err != nil {
		return nil
	}

	for i, cmd := range cmds {
		if result, err := cmd.Result(); err == nil {
			info[infoSections[i]] = []string{}
			splitInfo := strings.Split(result, "\r\n")[1:]
			for _, item := range splitInfo {
				itemTrimmed := strings.TrimSpace(item)
				if len(itemTrimmed) > 0 {
					info[infoSections[i]] = append(info[infoSections[i]], itemTrimmed)
				}
			}
		}
	}

	return info
}

func retrieveSlaves(client *redis.Client, master string, pod *pod) {
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

func retrievePodList(client *redis.Client) map[string]*pod {
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
		l2n.Printf("Error retrieving POD list: %s", err.Error())
		return nil
	}

	return podList
}

func getSentinelClient(sentinels []string) (*redis.Client, error) {
	var client *redis.Client
	foundValidSentinel := false

	for i := 0; i < len(sentinels) && foundValidSentinel == false; i++ {
		client = redis.NewClient(&redis.Options{
			Addr: sentinels[i],
		})

		if err := client.Ping().Err(); err == nil {
			foundValidSentinel = true
		}
	}

	if !foundValidSentinel {
		return nil, errors.New("No valid Sentinel found")
	}

	setRedisClientName(client)

	return client, nil
}

func apples(name string, addr string, master bool) {
	var role string
	if master {
		role = "master"
	} else {
		role = "slave"
	}

	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	if err := client.Ping().Err(); err != nil {
		l2n.Printf("Error connecting to %s '%s' of pod %s\n", role, addr, name)
		pd("Error connecting to "+role, map[string]interface{}{
			"Pod":     name,
			"Address": addr,
		})
		return
	}

	setRedisClientName(client)

	l2n.Printf("Connected to Redis '%s' @ %s", name, addr)

	ticker := time.NewTicker(time.Second * 30)
	for range ticker.C {
		blob := getInfoAsBlob(client)
		for _, section := range infoSections {
			dispatcher.Publish("redis-"+role, addr, section, blob[section])
		}
	}

	l2n.Printf("Shouldn't be here - Stopped checking %s", name)
}

// StartRedisMonitoring starts the monitoring of the buns
func StartRedisMonitoring(sentinels []string) {
	if sentinel, err := getSentinelClient(sentinels); err == nil {
		masterPodList = retrievePodList(sentinel)
		l2n.Printf("Current POD list: %+v", masterPodList)

		for k, v := range masterPodList {
			retrieveSlaves(sentinel, k, v)
			go apples(k, v.Addr, true)
			for _, addr := range v.Slaves {
				go apples(k, addr, false)
			}
		}
	}
}

func init() {}
