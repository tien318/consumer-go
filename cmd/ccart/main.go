package main

import (
	"database/sql"
	"os"

	"bitbucket.org/vunv92/consumer/config"
	"bitbucket.org/vunv92/consumer/console"
	"bitbucket.org/vunv92/consumer/mysql"
	"bitbucket.org/vunv92/consumer/redis"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	initLog()

	config.Load()

	redis.Init()

}

func main() {
	// Connect to database
	db, err := sql.Open("mysql", viper.GetString("mysql.dns"))

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	// Create services
	appShopService := &mysql.AppShopService{DB: db}
	shopService := &mysql.ShopService{DB: db}

	// Command
	cmd := &console.Command{
		AppShopService: appShopService,
		ShopService:    shopService,
	}

	cmd.Schedule()
	select {}

	// cmd.BuildJSONStatisticFile()
}

func initLog() {
	logFile, err := os.OpenFile("/var/log/ccart.log", os.O_CREATE|os.O_WRONLY, 0666)

	if err == nil {
		log.SetOutput(logFile)
	} else {
		log.Fatal(err)
		log.Info("Failed to log to file, using default stderr")
	}
}
