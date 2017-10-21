package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"beeketing.com/beeketing-consumer-go/config"
	"beeketing.com/beeketing-consumer-go/mysql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

const (
	appCode = "pusher"
)

var (
	// appShopService      *mysql.AppShopService
	// shopService         *mysql.ShopService
	// keyValueService     *mysql.KeyValueSettingService
	subscriptionService *mysql.SubscriptionService
	settingService      *mysql.SettingService
	// orderService        *mongo.OrderService
	// productService      *mongo.ProductService
)

func init() {
	config.Load()
}

func main() {
	log.Println("Start Pusher Consumer")

	// Connect to mysql
	db, err := sql.Open("mysql", viper.GetString("mysql.dsn"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.SetConnMaxLifetime(time.Second * 20)
	db.SetMaxIdleConns(30)
	db.SetMaxOpenConns(30)

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	subscriptionService = &mysql.SubscriptionService{DB: db}
	settingService = &mysql.SettingService{DB: db}

	conn, err := amqp.Dial(viper.GetString("rabbitmq.url"))
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"pusher_handle_abandoned_cart", // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			go handleMessage(d.Body)
		}
	}()

	log.Info("[*] Waiting for messages")
	<-forever
}

func handleMessage(message []byte) {
	log.Infof("Received a message: %s", message)

	var msgData map[string]interface{}
	err := json.Unmarshal(message, &msgData)
	if err != nil {
		log.Errorf("%s: %s", "Unmarchal message failed", err)
		return
	}

	if _, ok := msgData["shop_id"]; !ok {
		log.Errorf("%s", "shop_id not found")
		return
	}

	if _, ok := msgData["cart_token"]; !ok {
		log.Errorf("%s", "cart_token not found")
		return
	}

	if _, ok := msgData["created_at"]; !ok {
		log.Errorf("%s", "created_at not found")
		return
	}

	createdAt, err := time.Parse(time.RFC3339, msgData["created_at"].(string))
	if err != nil {
		log.Errorf("%s: %s", "parse created_at failed", err)
	}

	shopID := int64(msgData["shop_id"].(float64))

	cartToken := msgData["cart_token"].(string)

	sub, err := subscriptionService.GetByCartToken(cartToken)
	if err != nil && err != sql.ErrNoRows {
		log.Errorf("%s: %s", "get sub failed", err)
	}

	if err == sql.ErrNoRows {
		log.Info("No subscription found with cart token: ", cartToken)
		return
	}

	// fmt.Println(cartToken)
	// fmt.Println(shopID)
	fmt.Println(createdAt)
	fmt.Println(sub.Subscription)

	keyStrings := [3]string{
		"pusher_cart_reminder_15_mins",
		"pusher_cart_reminder_1_hour",
		"pusher_cart_reminder_24_hours",
	}

	for _, keyString := range keyStrings {
		log.Info(keyString)

		setting, err := settingService.Get(shopID, appCode, keyString)

		if err != nil {
			log.Errorf("%s: %s", "get setting failed", err)
		}

		var settingData map[string]interface{}
		err = json.Unmarshal([]byte(setting.Value), &settingData)
		if err != nil {
			log.Errorf("%s: %s", "Unmarshal message failed", err)
			continue
		}

		if _, ok := settingData["enable"]; !ok {
			log.Error("enable not found")
			return
		}

		if !settingData["enable"].(bool) {
			log.Info("Disabled")
			return
		}

		log.Info(settingData)
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
