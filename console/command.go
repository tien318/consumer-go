package console

import (
	"encoding/base64"
	"io/ioutil"
	"strconv"
	"strings"

	"encoding/json"

	"bitbucket.org/vunv92/consumer"
	"bitbucket.org/vunv92/consumer/redis"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
)

// Command lorem
type Command struct {
	AppShopService consumer.AppShopService
	ShopService    consumer.ShopService
}

// Schedule lorem
func (c *Command) Schedule() {
	cron := cron.New()

	log.Info("run BuildJSONStatisticFile every 10s")
	cron.AddFunc("@every 10s", func() {
		c.BuildJSONStatisticFile()
	})

	cron.Start()
}

// BuildJSONStatisticFile lorem
func (c *Command) BuildJSONStatisticFile() {
	log.Info("BuildJSONStatisticFile")
	appShops, err := c.AppShopService.GetByAppID(16)

	if err != nil {
		log.Fatal(err)
	}

	for _, appShop := range appShops {
		c.BuildShopStatisticJSONFile(appShop)
	}
}

// BuildShopStatisticJSONFile lorem
func (c *Command) BuildShopStatisticJSONFile(appShop *consumer.AppShop) {

	// get shop from mysql
	shop, err := c.ShopService.GetByID(appShop.ShopID)
	if err != nil {
		log.Fatal(err)
	}
	if shop == nil {
		return
	}

	// get statistics data from redis
	hashName := strings.Join([]string{"ps:", strconv.Itoa(appShop.ShopID)}, "")
	hash, err := redis.Client.HGetAll(hashName).Result()
	if err != nil {
		log.Fatal(err)
	}

	if len(hash) > 0 {
		log.Info("Shop ", appShop.ShopID, " | has ", len(hash), " statistics record")
	} else {
		log.Error("Shop ", appShop.ShopID, " | has no statistics record")
		return
	}

	stats := make(map[string][]int)

	for key, val := range hash {
		fields := strings.Split(key, ":")
		if len(fields) != 2 {
			log.Fatal("Invalid hash field")
		}

		if _, ok := stats[fields[0]]; !ok {
			stats[fields[0]] = []int{0, 0, 0}
		}

		if fields[1] == "v" {
			stats[fields[0]][0], _ = strconv.Atoi(val)
		} else if fields[1] == "ac" {
			stats[fields[0]][1], _ = strconv.Atoi(val)
		} else if fields[1] == "p" {
			stats[fields[0]][2], _ = strconv.Atoi(val)
		}
	}

	// write statistics data to json file
	statStr, _ := json.Marshal(stats)
	fileName := base64.StdEncoding.EncodeToString([]byte(shop.APIKey))
	fileName = strings.Join([]string{fileName, "json"}, ".")

	err = ioutil.WriteFile(fileName, statStr, 0777)
	if err != nil {
		log.Fatal(err)
	}

	log.Info(string(statStr))
}
