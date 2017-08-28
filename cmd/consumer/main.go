package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"

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

	initLog()
}

func main() {
	log.Println("Start CCart Consumer")

	// Connect to mysql
	db, err := sql.Open("mysql", viper.GetString("mysql.dsn"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

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

	if _, err := os.Stat("rest"); os.IsNotExist(err) {
		os.Mkdir("rest", 0777)
	}

	forever := make(chan bool)

	go func() {
		var msgData map[string]int

		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)

			err := json.Unmarshal(d.Body, &msgData)
			if err != nil {
				log.Fatal(err)
			}

			appShop, err := appShopService.GetByID(msgData["app_shop_id"])
			if err != nil {
				log.Fatal(err)
			}

			shop, err := shopService.GetByID(appShop.ShopID)
			if err != nil {
				log.Fatal(err)
			}
			log.Println("shop id:", shop.ID)
			log.Println("api key:", shop.APIKey)

			stats := make(map[string][]int)

			products, err := productService.GetByShopID(appShop.ShopID)
			if err != nil {
				log.Fatal(err)
			}

			for _, product := range products {
				stats[strconv.Itoa(product.RefID)] = productService.GetDefaultStatisticsData(product.RefID)
			}

			statStr, _ := json.Marshal(stats)

			fileName := base64.StdEncoding.EncodeToString([]byte(shop.APIKey))
			filePath := "rest/" + fileName + ".json"

			log.Println("write to:", filePath)

			err = ioutil.WriteFile(filePath, statStr, 0777)
			if err != nil {
				log.Fatal(err)
			}
		}
	}()

	log.Printf("[*] Waiting for messages")
	<-forever
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func initLog() {
	// log.SetFormatter(&log.JSONFormatter{})

	logOutput := viper.GetString("log.output")

	if logOutput == "file" {
		logFile, err := os.OpenFile("ccart.log", os.O_CREATE|os.O_WRONLY, 0666)

		if err == nil {
			log.SetOutput(logFile)
		} else {
			log.Fatal(err)
			log.Info("Failed to log to file, using default stderr")
		}
	}
}
