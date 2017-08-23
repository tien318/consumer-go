package console

import (
	"encoding/base64"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"encoding/json"

	"beeketing.com/consumer"
	"beeketing.com/consumer/redis"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var apiKeys = make(map[int]string)

const ccartAppID int = 16

// Command lorem
type Command struct {
	AppShopService consumer.AppShopService
	ShopService    consumer.ShopService
	OrderService   consumer.OrderService
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
	// get all app shop
	appShops, err := c.AppShopService.GetByAppID(ccartAppID)

	if err != nil {
		log.Fatal(err)
	}

	shopIDs := []int{}
	for _, appShop := range appShops {
		shopIDs = append(shopIDs, appShop.ShopID)
	}

	// get shops to get apikeys
	shops, err := c.ShopService.GetByIDs(shopIDs)
	for _, shop := range shops {
		apiKeys[shop.ID] = shop.APIKey
	}

	// create rest file if not exist
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
	hashName := "ps:" + strconv.Itoa(appShop.ShopID)
	hash, err := redis.Client.HGetAll(hashName).Result()
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Shop ", appShop.ShopID, " | has ", len(hash), " statistics record")
	if len(hash) == 0 {
		return
	}

	stats := make(map[string][]int)

	// statistic file
	fileName := base64.StdEncoding.EncodeToString([]byte(apiKeys[appShop.ShopID]))
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
			stats[fields[0]] = c.initProductStatisticsData(fields[0])
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

	// log.Info(string(statStr))
}

func (c *Command) initProductStatisticsData(productID string) []int {
	var view, addToCart, purchase int = 0, 0, 0

	// query to mongo to get count order
	id, _ := strconv.Atoi(productID)
	purchase, err := c.OrderService.CountByProductRefID(id)

	if err != nil {
		log.Fatal(err)
	}

	addToCart = purchase * 2
	view = addToCart * 5

	return []int{view, addToCart, purchase}
}
