package console

import (
	"encoding/base64"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"encoding/json"

	"bitbucket.org/vunv92/consumer"
	"bitbucket.org/vunv92/consumer/redis"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var apiKeys = make(map[int]string)

// Command lorem
type Command struct {
	AppShopService consumer.AppShopService
	ShopService    consumer.ShopService
}

// Schedule lorem
func (c *Command) Schedule() {
	cron := cron.New()

	ccartInterval := viper.GetString("crons.ccart")
	log.Info("run BuildJSONStatisticFile every ", ccartInterval)
	cron.AddFunc("@every "+ccartInterval, func() {
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

	if _, err := os.Stat("rest"); os.IsNotExist(err) {
		os.Mkdir("rest", 0777)
	}

	for _, appShop := range appShops {
		go c.BuildShopStatisticJSONFile(appShop)
	}
}

// BuildShopStatisticJSONFile lorem
func (c *Command) BuildShopStatisticJSONFile(appShop *consumer.AppShop) {
	// get statistics data from redis
	hashName := strings.Join([]string{"ps:", strconv.Itoa(appShop.ShopID)}, "")
	hash, err := redis.Client.HGetAll(hashName).Result()
	if err != nil {
		log.Fatal(err)
	}

	if len(hash) > 0 {
		log.Info("Shop ", appShop.ShopID, " | has ", len(hash), " statistics record")
	} else {
		return
	}

	apiKey := ""

	if _, ok := apiKeys[appShop.ShopID]; !ok {
		// get shop from mysql
		shop, err := c.ShopService.GetByID(appShop.ShopID)
		if err != nil {
			log.Fatal(err)
		}
		if shop == nil {
			return
		}

		apiKeys[appShop.ShopID] = shop.APIKey
		apiKey = shop.APIKey
	} else {
		apiKey = apiKeys[appShop.ShopID]
	}

	stats := make(map[string][]int)

	// statistic file
	fileName := base64.StdEncoding.EncodeToString([]byte(apiKey))
	filePath := "rest/" + fileName + ".json"

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Fatal(err)
		}

		if err := json.Unmarshal(data, &stats); err != nil {
			log.Fatal(err)
		}
	}

	for key, val := range hash {
		// fields[0]: product id
		// fields[1]: action type
		fields := strings.Split(key, ":")

		if len(fields) != 2 {
			log.Fatal("Invalid hash field")
		}

		if _, ok := stats[fields[0]]; !ok {
			stats[fields[0]] = []int{0, 0, 0}
		}

		count, _ := strconv.Atoi(val)

		if fields[1] == "v" {
			stats[fields[0]][0] += count
		} else if fields[1] == "ac" {
			stats[fields[0]][1] += count
		} else if fields[1] == "p" {
			stats[fields[0]][2] += count
		}
	}

	// Delete hash
	_, err = redis.Client.Del(hashName).Result()
	if err != nil {
		log.Fatal(err)
	}

	// write statistics data to json file
	statStr, _ := json.Marshal(stats)
	err = ioutil.WriteFile(filePath, statStr, 0777)
	if err != nil {
		log.Fatal(err)
	}

	log.Info(string(statStr))
}
