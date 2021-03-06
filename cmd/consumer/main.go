package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"strconv"

	mgo "gopkg.in/mgo.v2"

	consumer "beeketing.com/beeketing-consumer-go"
	"beeketing.com/beeketing-consumer-go/mongo"

	log "github.com/sirupsen/logrus"

	"beeketing.com/beeketing-consumer-go/config"
	"beeketing.com/beeketing-consumer-go/mysql"
	"beeketing.com/beeketing-consumer-go/statistic"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

var (
	appShopService  consumer.AppShopService
	shopService     consumer.ShopService
	keyValueService consumer.KeyValueSettingService
	orderService    consumer.OrderService
	productService  consumer.ProductService
)

func init() {
	config.Load()
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

	appShopService = &mysql.AppShopService{DB: db}
	shopService = &mysql.ShopService{DB: db}
	keyValueService = &mysql.KeyValueSettingService{DB: db}
	orderService = mongo.NewOrderService(session)
	productService = mongo.NewProductService(session, orderService)

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
		os.Mkdir(restPath, 055)
	}

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

	var msgData map[string]int
	err := json.Unmarshal(message, &msgData)
	if err != nil {
		log.Errorf("%s: %s", "Unmarchal message failed", err)
		return
	}

	if _, ok := msgData["app_shop_id"]; !ok {
		log.Errorf("%s", "app_shop_id not found")
		return
	}

	// check init
	keyName := fmt.Sprintf("ccart_init_data_%d", msgData["app_shop_id"])
	_, err = keyValueService.GetByKeyName(keyName)
	if err == nil {
		log.Infof("%d | key value existed", msgData["app_shop_id"])
		return
	}

	err = keyValueService.CreateKeyValue(keyName, "1")
	if err != nil {
		log.Errorf("%s: %s", "Create key value failed", err)
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

	log.Infof("%d |Count Product: %d", msgData["app_shop_id"], len(products))

	productStat := statistic.NewProductStat()
	for _, product := range products {
		productStat.Data[strconv.Itoa(product.RefID)] = productService.GetDefaultStatisticsData(shop.ID, product.RefID)
	}

	fileName := base64.StdEncoding.EncodeToString([]byte(shop.APIKey))
	filePath := viper.GetString("static.path") + "/rest/" + fileName + ".json"

	err = ioutil.WriteFile(filePath, productStat.GetJSONData(), 0777)
	if err != nil {
		log.Errorf("%s: %s", "Write json data to file failed", err)
	}

	log.Infof("%d | Create default data: %s", msgData["app_shop_id"], fileName)
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
