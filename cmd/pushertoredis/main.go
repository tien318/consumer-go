package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	mgo "gopkg.in/mgo.v2"

	consumer "beeketing.com/beeketing-consumer-go"
	"beeketing.com/beeketing-consumer-go/config"
	"beeketing.com/beeketing-consumer-go/mongo"
	"beeketing.com/beeketing-consumer-go/mysql"
	"beeketing.com/beeketing-consumer-go/redis"
)

var (
	appService     consumer.AppService
	appShopService consumer.AppShopService
	cartService    consumer.CartService
)

func init() {
	config.Load()

	redis.Init()
}

func main() {
	fmt.Println("Pusher consumer to Redis")

	db, err := mysql.NewMysql()
	defer db.Close()

	if err != nil {
		panic(err)
	}

	appService = &mysql.AppService{DB: db}
	appShopService = &mysql.AppShopService{DB: db}

	// run()

	c := cron.New()

	log.Info("run every 5m")
	c.AddFunc("@every 5m", func() {
		run()
	})

	c.Start()

	select {}
}

func run() {
	session, err := mgo.Dial(viper.GetString("mongodb.url"))

	if err != nil {
		log.Fatalf("%s: %s", "Failed to connect to mongo", err)
	}
	cartService = mongo.NewCartService(session)

	// init time
	updatedAtMin := time.Now().Add(-time.Minute * 20).Format(time.RFC3339)
	updatedAtMax := time.Now().Add(-time.Minute * 15).Format(time.RFC3339)

	carts, err := cartService.GetShopIDOfAbandonedCarts(updatedAtMin, updatedAtMax)
	if err != nil {
		log.Errorf("%s: %s", "Get abandoned carts failed", err)
		return
	}

	app, err := appService.GetByAppCode("pusher")
	if err != nil {
		log.Errorf("%s: %s", "Get App failed", err)
		return
	}

	for _, cart := range carts {
		_, err := appShopService.GetByShopIDAndAppID(cart.ShopID, app.ID)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Info("Shop does not install pusher")
				continue
			}
			log.Errorf("Fail to get app shop")
			return
		}

		// TODU: change key?
		shopStr := strconv.Itoa(cart.ShopID)
		key := fmt.Sprintf("%v:%v", shopStr, cart.CartToken)
		// TODU: remove
		log.Infof("key: %v", key)

		err = redis.Client.HSet("pusher", key, cart.CartToken).Err()
		if err != nil {
			log.Errorf("Fail to save cart token to redis, detail: %v", err)
			return
		}
		value, err := redis.Client.HGet("pusher", key).Result()
		if err != nil {
			log.Errorf("Fail to get cart token to redis, , detail: %v", err)
			return
		}
		// TODU: remove
		log.Infof("value: %v", value)
	}
}
