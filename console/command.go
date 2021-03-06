package console

import (
	"encoding/base64"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"encoding/json"

	"beeketing.com/beeketing-consumer-go"
	"beeketing.com/beeketing-consumer-go/redis"
	"beeketing.com/beeketing-consumer-go/statistic"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var apiKeys = make(map[int]string)

const ccartAppCode string = "countdown_cart"

// Command lorem
type Command struct {
	AppService             consumer.AppService
	AppShopService         consumer.AppShopService
	ShopService            consumer.ShopService
	OrderService           consumer.OrderService
	ProductService         consumer.ProductService
	KeyValueSettingService consumer.KeyValueSettingService
}

// Schedule lorem
func (c *Command) Schedule() {
	cron := cron.New()

	ccartInterval := viper.GetString("crons.ccart")
	log.Info("run BuildJSONStatisticFile every ", ccartInterval)
	cron.AddFunc("@every "+ccartInterval, func() {
		c.BuildJSONStatisticFile()
	})

	precInterval := viper.GetString("crons.prec")
	log.Info("run UpdatePrecStat every ", precInterval)
	cron.AddFunc("@every "+precInterval, func() {
		c.UpdatePrecStats()
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
	// log.Info("Count ShopIDs: ", len(shopIDs))

	// get shops to get apikeys
	shops, err := c.ShopService.GetByIDs(shopIDs)
	if err != nil {
		log.Errorf("%s: %s", "Get shops by ids failed", err)
	}
	// log.Info("Count shops: ", len(shops))

	for _, shop := range shops {
		apiKeys[shop.ID] = shop.APIKey
	}

	// create rest file if not exist
	restPath := viper.GetString("static.path") + "/rest/"
	if _, err := os.Stat(restPath); os.IsNotExist(err) {
		os.Mkdir(restPath, 0755)
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

	productStat := statistic.NewProductStat()

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

		if err := json.Unmarshal(data, &productStat); err != nil {
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

		if _, ok := productStat.Data[fields[0]]; !ok {
			refID, _ := strconv.Atoi(fields[0])

			productStat.Data[fields[0]] = c.ProductService.GetDefaultStatisticsData(appShop.ShopID, refID)
		}

		count, _ := strconv.Atoi(val)

		if fields[1] == "v" {
			productStat.Data[fields[0]][0] += count
		} else if fields[1] == "ac" {
			productStat.Data[fields[0]][1] += count
		} else if fields[1] == "p" {
			productStat.Data[fields[0]][2] += count
		}
	}

	// Delete hash
	_, err = redis.Client.Del(hashName).Result()
	if err != nil {
		log.Errorf("%s: %s | %s", "Delete hash failed", hashName, err)
		return
	}

	// write statistics data to json file
	err = ioutil.WriteFile(filePath, productStat.GetJSONData(), 0777)
	if err != nil {
		log.Errorf("%s: %s", "Wrire Json to file failed", err)
		return
	}
}
