package main

import (
	"database/sql"
	"time"

	"beeketing.com/beeketing-consumer-go/config"
	"beeketing.com/beeketing-consumer-go/console"
	"beeketing.com/beeketing-consumer-go/mongo"
	"beeketing.com/beeketing-consumer-go/mysql"
	"beeketing.com/beeketing-consumer-go/redis"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	mgo "gopkg.in/mgo.v2"
)

func init() {
	config.Load()

	// initLog()

	redis.Init()
}

func main() {
	log.Info("Start Ccart Consumer")
	// Connect to mysql
	db, err := sql.Open("mysql", viper.GetString("mysql.dsn"))

	if err != nil {
		log.Fatalf("%s: %s", "Failed to connect to mysql", err)
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
		log.Fatalf("%s: %s", "Failed to connect to mongo", err)
	}
	defer session.Close()

	// Create services
	appService := &mysql.AppService{DB: db}
	keyValueSettingService := &mysql.KeyValueSettingService{DB: db}
	appShopService := &mysql.AppShopService{DB: db}
	shopService := &mysql.ShopService{DB: db}
	orderService := mongo.NewOrderService(session)
	productService := mongo.NewProductService(session, orderService)

	// Command
	cmd := &console.Command{
		AppService:             appService,
		AppShopService:         appShopService,
		ShopService:            shopService,
		OrderService:           orderService,
		ProductService:         productService,
		KeyValueSettingService: keyValueSettingService,
	}

	cmd.Schedule()

	// for development only
	// cmd.BuildJSONStatisticFile()

	select {}
}
