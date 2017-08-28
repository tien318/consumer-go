package main

import (
	"database/sql"

	"beeketing.com/consumer/config"
	"beeketing.com/consumer/console"
	"beeketing.com/consumer/mongo"
	"beeketing.com/consumer/mysql"
	"beeketing.com/consumer/redis"

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
	appShopService := &mysql.AppShopService{DB: db}
	shopService := &mysql.ShopService{DB: db}
	orderService := &mongo.OrderService{Session: session}

	// Command
	cmd := &console.Command{
		AppShopService: appShopService,
		ShopService:    shopService,
		OrderService:   orderService,
	}

	cmd.Schedule()

	// for development only
	// cmd.BuildJSONStatisticFile()

	select {}
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
