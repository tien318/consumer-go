package console

import (
	"encoding/json"
	"strconv"
	"beeketing.com/beeketing-consumer-go"
	"beeketing.com/beeketing-consumer-go/redis"
	log "github.com/sirupsen/logrus"
)

const precAppCode string = "precommend"

// UpdatePrecStats lorem
func (c *Command) UpdatePrecStats() {
	log.Info("Build Prec Statistic")
	app, err := c.AppService.GetByAppCode(precAppCode)

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

	for _, appShop := range appShops {
		go c.UpdatePrecStat(appShop)
	}
}


// UpdatePrecStat lorem
func (c *Command) UpdatePrecStat(appShop *consumer.AppShop) {
	// get statistics data from redis
	hashName := "mcp:" + strconv.Itoa(appShop.ShopID)
	hash, err := redis.Client.HGetAll(hashName).Result()
	if err != nil {
		log.Errorf("%s: %s", "Get prec hash from redis failed", err)
		return
	}

	// log.Info("Shop ", appShop.ShopID, " | has ", len(hash), " statistics record")
	if len(hash) == 0 {
		return
	}

	keyName := "prec_product_clicked_" + strconv.Itoa(appShop.ShopID)
	var productStat = make(map[int]int)
	data, err := c.KeyValueSettingService.GetByKeyName(keyName)

	//if err != nil {
	//	log.Errorf("%s: %s | %s", "Read prec json failed", appShop.ShopID, err)
	//	return
	//}

	if data == nil {
		json.Unmarshal([]byte("{}"), &productStat)
	} else {
		if err := json.Unmarshal([]byte(data.KeyValue), &productStat); err != nil {
			log.Errorf("%s: %s | %s", "Unmarshal prec json failed", appShop.ShopID, err)
			return
		}
	}

	for key, val := range hash {
		productId, err := strconv.Atoi(key)
		if err != nil {
			continue
		}

		if _, ok := productStat[productId]; !ok {
			productStat[productId] = 0
		}

		count, _ := strconv.Atoi(val)

		productStat[productId] += count
	}

	// write statistics data to json file
	jsonData, err := json.Marshal(productStat)
	if err != nil {
		log.Errorf("%s: %s", "Wrire Json to file failed", err)
		return
	} else {
		if data != nil {
			c.KeyValueSettingService.UpdateKeyValue(data.ID, string(jsonData))
		} else {
			c.KeyValueSettingService.CreateKeyValue(keyName, string(jsonData))
		}
	}

	// Delete hash
	_, err = redis.Client.Del(hashName).Result()

	if err != nil {
		log.Errorf("%s: %s | %s", "Delete prec hash failed", hashName, err)
		return
	}
}