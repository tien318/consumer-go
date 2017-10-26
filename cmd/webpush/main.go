package main

import (
	"time"

	consumer "beeketing.com/beeketing-consumer-go"
	"beeketing.com/beeketing-consumer-go/config"
	"beeketing.com/beeketing-consumer-go/mongo"
	"beeketing.com/beeketing-consumer-go/mysql"
	"beeketing.com/beeketing-consumer-go/redis"
	"beeketing.com/beeketing-consumer-go/webpush"
	goredis "github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	mgo "gopkg.in/mgo.v2"
)

const (
	cacheKey         = "pusher_statistic_cache_key"
	startStatisticID = 3000000000
)

var (
	notificationService consumer.WebNotificationService
	statisticService    consumer.StatisticService
)

func init() {
	config.Load()

	redis.InitPersistentRedis()
}

func main() {
	db, err := mysql.NewMysql()
	defer db.Close()

	if err != nil {
		panic(err)
	}

	// mongodb
	session, err := mgo.Dial(viper.GetString("mongodb.url"))
	if err != nil {
		log.Fatalf("%s: %s", "Failed to connect to mongo", err)
	}
	defer session.Close()

	notificationService = &mysql.WebNotificationService{DB: db}
	statisticService = mongo.NewStatisticService(session)

	c := cron.New()

	log.Info("Run Send notification every 1m")

	c.AddFunc("@every 1m", func() {
		run()
	})

	c.Start()

	// run()

	select {}
}

func run() {
	notifications, err := notificationService.GetNotificationToSend()
	if err != nil {
		log.Errorf("%s: %s", "Get notification failed", err)
		return
	}

	log.Info("Count Notification: ", len(notifications))

	for _, notification := range notifications {
		go send(notification)
	}
}

func send(noti *consumer.WebNotification) {
	notificationService.UpdateSent(noti)

	webpush.Send(noti.Subscription, noti.Data)

	updateStatistic(noti.ShopID, "day")
	updateStatistic(noti.ShopID, "total")
}

func updateStatistic(shopID int64, timeType string) {
	// params
	statisticType := "shop"
	t := time.Now()
	year, month, day := t.Date()
	time := time.Date(year, month, day, 0, 0, 0, 0, t.Location())

	// get statistic by params
	stat, err := statisticService.Get(shopID, statisticType, shopID, timeType, time)
	if err != nil && err != mgo.ErrNotFound {
		log.Fatal(err)
	}

	if err == mgo.ErrNotFound {
		// if stat not found, insert
		stat = &consumer.Statistic{
			ID:       getNewStatisticID(),
			ShopID:   shopID,
			RefID:    shopID,
			Type:     statisticType,
			Data:     make(map[string]int64),
			TimeType: timeType,
			Time:     time.Unix(),
		}

		err = statisticService.Add(stat)

		if err != nil {
			log.Fatal(err)
		}
	}

	// increase value
	err = statisticService.Increase(stat, "pusher_count", 1)

	if err != nil {
		log.Fatal(err)
	}
}

func getNewStatisticID() int64 {
	id, err := redis.ClientPersistent.Get(cacheKey).Int64()

	if err == goredis.Nil {
		id = startStatisticID
	} else if err != nil {
		log.Fatal(err)
	} else {
		id++
	}

	err = redis.ClientPersistent.Set(cacheKey, id, 0).Err()

	if err != nil {
		log.Fatal(err)
	}

	return id
}
