package console

import (
	"encoding/base64"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"encoding/json"

	"beeketing.com/consumer"
	"beeketing.com/consumer/redis"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var apiKeys = make(map[int]string)

const ccartAppCode string = "countdown_cart"

// Command lorem
type Command struct {
	AppService     consumer.AppService
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
	log.Info("Build JSON Statistic File")
	app, err := c.AppService.GetByAppCode(ccartAppCode)

	if err != nil {
		log.Errorf("%s: %s", "Get App failed", err)
	}

	// get all app shop
	appShops, err := c.AppShopService.GetByAppID(app.ID)

	if err != nil {
		log.Errorf("%s: %s", "Get appShops failed", err)
		return
	}

	if len(appShops) == 0 {
		return
	}

	shopIDs := []int{}
	for _, appShop := range appShops {
		shopIDs = append(shopIDs, appShop.ShopID)
	}
	log.Info("Count ShopIDs: ", len(shopIDs))

	// get shops to get apikeys
	shops, err := c.ShopService.GetByIDs(shopIDs)
	for _, shop := range shops {
		apiKeys[shop.ID] = shop.APIKey
	}

	log.Info("Count shops: ", len(shops))

	// create rest file if not exist
	restPath := viper.GetString("static.path") + "/rest/"
	if _, err := os.Stat(restPath); os.IsNotExist(err) {
		os.Mkdir(restPath, 0777)
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
		log.Errorf("%s: %s", "Get hash from redis failed", err)
		return
	}

	// log.Info("Shop ", appShop.ShopID, " | has ", len(hash), " statistics record")
	if len(hash) == 0 {
		return
	}

	stats := make(map[string][]int)

	if _, ok := apiKeys[appShop.ShopID]; !ok {
		return
	}

	// statistic file
	fileName := base64.StdEncoding.EncodeToString([]byte(apiKeys[appShop.ShopID]))
	filePath := viper.GetString("static.path") + "/rest/" + fileName + ".json"

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Errorf("%s: %s | %s", "Read json file failed", filePath, err)
			return
		}

		if err := json.Unmarshal(data, &stats); err != nil {
			log.Errorf("%s: %s | %s", "Unmarshal json file failed", filePath, err)
			return
		}
	}

	for key, val := range hash {
		// fields[0]: product id
		// fields[1]: action type
		fields := strings.Split(key, ":")

		if len(fields) != 2 {
			log.Errorf("%s: %s | %s", "Invalid hash field", key, err)
			return
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
		log.Errorf("%s: %s | %s", "Delete hash failed", hashName, err)
		return
	}

	// write statistics data to json file
	statStr, _ := json.Marshal(stats)
	err = ioutil.WriteFile(filePath, statStr, 0777)
	if err != nil {
		log.Errorf("%s: %s", "Wrire Json to file failed", err)
		return
	}

	// log.Info(string(statStr))
}

func (c *Command) initProductStatisticsData(productID string) []int {
	var view, addToCart, purchase int = 0, 0, 0

	// query to mongo to get count order
	id, _ := strconv.Atoi(productID)
	purchase, err := c.OrderService.CountByProductRefID(id)

	if err != nil {
		log.Error(err)
		return []int{view, addToCart, purchase}
	}

	rand.Seed(time.Now().UnixNano())
	addToCart = int(float32(purchase) * (rand.Float32() + 1))

	view = addToCart * (rand.Intn(10) + 10)

	return []int{view, addToCart, purchase}
}
