package main

import (
	"database/sql"
	"fmt"
	"os"

	"net/http"

	"beeketing.com/consumer/config"
	"beeketing.com/consumer/console"
	"beeketing.com/consumer/mongo"
	"beeketing.com/consumer/mysql"
	"beeketing.com/consumer/redis"

	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	mgo "gopkg.in/mgo.v2"
)

func init() {
	config.Load()

	initLog()

	redis.Init()
}

func main() {
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

	router := httprouter.New()

	router.GET("/", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		fmt.Fprintln(w, "Hello")
	})

	router.ServeFiles("/rest/*filepath", http.Dir("./rest"))

	log.Info("Server listen and serve at http://localhost:8088")

	log.Fatal(http.ListenAndServe(":8088", router))

	// select {}
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
