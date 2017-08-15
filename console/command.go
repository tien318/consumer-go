package console

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"encoding/json"

	"github.com/nvvu/consumer"
	"github.com/nvvu/consumer/redis"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
)

// Command lorem
type Command struct {
	Ass consumer.AppShopService
}

// Schedule lorem
func (c *Command) Schedule() {
	cron := cron.New()

	cron.AddFunc("@every 5s", func() {
		c.BuildJSONStatisticFile()
	})

	cron.Start()
}

// BuildJSONStatisticFile lorem
func (c *Command) BuildJSONStatisticFile() {
	log.Info("BuildJSONStatisticFile")
	// get all shop using ccart
	appShops, err := c.Ass.GetByAppID(16)

	if err != nil {
		log.Fatal(err)
	}

	for _, as := range appShops {
		c.BuildShopStatisticJSONFile(as)
	}
}

// BuildShopStatisticJSONFile lorem
func (c *Command) BuildShopStatisticJSONFile(as *consumer.AppShop) {
	hashName := strings.Join([]string{"ps:", strconv.Itoa(as.ShopID)}, "")

	hash, err := redis.Client.HGetAll(hashName).Result()

	if err != nil {
		log.Fatal(err)
	}

	if len(hash) > 0 {
		log.Info("Shop ", as.ShopID, " | has ", len(hash), " statistics record")
	} else {
		log.Error("Shop ", as.ShopID, " | has no statistics record")
		return
	}

	stats := make(map[string][]int)

	for key, val := range hash {

		fields := strings.Split(key, ":")
		if len(fields) != 2 {
			log.Fatal("Invalid hash field")
		}

		if _, ok := stats[fields[0]]; ok {
			if fields[1] == "v" {
				stats[fields[0]][0], _ = strconv.Atoi(val)
			} else if fields[1] == "ac" {
				stats[fields[0]][1], _ = strconv.Atoi(val)
			} else if fields[1] == "p" {
				stats[fields[0]][2], _ = strconv.Atoi(val)
			}
		} else {
			stats[fields[0]] = []int{0, 0, 0}
		}
	}

	statStr, _ := json.Marshal(stats)

	err = ioutil.WriteFile("stat.json", statStr, 0777)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(statStr))
}
