package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	"strconv"

	mgo "gopkg.in/mgo.v2"

	"beeketing.com/consumer/mongo"

	log "github.com/sirupsen/logrus"

	"beeketing.com/consumer/config"
	"beeketing.com/consumer/mysql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

func init() {
	config.Load()

	// initLog()
}

func main() {
	log.Println("Start CCart Consumer")

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

	// mongodb
	session, err := mgo.Dial(viper.GetString("mongodb.url"))
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	appShopService := &mysql.AppShopService{DB: db}
	shopService := &mysql.ShopService{DB: db}
	orderService := &mongo.OrderService{Session: session}
	productService := &mongo.ProductService{Session: session, OrderService: orderService}

	conn, err := amqp.Dial(viper.GetString("rabbitmq.url"))
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"countdown_cart_create_default_data", // name
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

	restPath := viper.GetString("static.path") + "/rest"
	if _, err := os.Stat(restPath); os.IsNotExist(err) {
		os.Mkdir(restPath, 0777)
	}

	forever := make(chan bool)

	go func() {
		var msgData map[string]int

		for d := range msgs {
			log.Infof("Received a message: %s", d.Body)

			err := json.Unmarshal(d.Body, &msgData)
			if err != nil {
				log.Errorf("%s: %s", "Unmarchal message failed", err)
				return
			}

			appShop, err := appShopService.GetByID(msgData["app_shop_id"])
			if err != nil {
				log.Errorf("%s: %s", "Get app shop failed", err)
				return
			}

			shop, err := shopService.GetByID(appShop.ShopID)
			if err != nil {
				log.Errorf("%s: %s", "Get shop failed", err)
				return
			}

			products, err := productService.GetByShopID(appShop.ShopID)
			if err != nil {
				log.Errorf("%s: %s", "Get products failed", err)
				return
			}

			stats := make(map[string][]int)
			for _, product := range products {
				stats[strconv.Itoa(product.RefID)] = productService.GetDefaultStatisticsData(product.RefID)
			}
			statStr, _ := json.Marshal(stats)

			fileName := base64.StdEncoding.EncodeToString([]byte(shop.APIKey))
			filePath := viper.GetString("static.path") + "/rest/" + fileName + ".json"

			err = ioutil.WriteFile(filePath, statStr, 0777)

			log.Info("Update json data:", fileName)

			if err != nil {
				log.Errorf("%s: %s", "Write json data to file failed", err)
			}
		}
	}()

	log.Info("[*] Waiting for messages")
	<-forever
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

// func initLog() {
// 	// log.SetFormatter(&log.JSONFormatter{})

// 	logOutput := viper.GetString("log.output")

// 	if logOutput == "file" {
// 		logFile, err := os.OpenFile("ccart.log", os.O_CREATE|os.O_WRONLY, 0666)

// 		if err == nil {
// 			log.SetOutput(logFile)
// 		} else {
// 			log.Fatal(err)
// 			log.Info("Failed to log to file, using default stderr")
// 		}
// 	}
// }
