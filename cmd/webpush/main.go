package main

import (
	consumer "beeketing.com/beeketing-consumer-go"
	"beeketing.com/beeketing-consumer-go/config"
	"beeketing.com/beeketing-consumer-go/mysql"
	"beeketing.com/beeketing-consumer-go/webpush"
	_ "github.com/go-sql-driver/mysql"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
)

func init() {
	config.Load()
}

var (
	notificationService consumer.WebNotificationService
)

func main() {
	db, err := mysql.NewMysql()
	defer db.Close()

	if err != nil {
		panic(err)
	}

	notificationService = &mysql.WebNotificationService{DB: db}

	c := cron.New()

	log.Info("Run Send notification every 1m")

	c.AddFunc("@every 5s", func() {
		run()
	})

	c.Start()

	select {}
}

func run() {
	log.Info("Send Notification")

	notifications, err := notificationService.GetNotificationToSend()
	if err != nil {
		log.Errorf("%s: %s", "Get notification failed", err)
		return
	}
	log.Info("Count Notification: ", len(notifications))

	for _, notification := range notifications {
		go webpush.Send(notification.Subscription, notification.Data)
	}
}
