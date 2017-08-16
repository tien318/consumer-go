package main

import (
	"database/sql"

	"bitbucket.org/vunv92/consumer/config"
	"bitbucket.org/vunv92/consumer/console"
	"bitbucket.org/vunv92/consumer/mysql"
	"bitbucket.org/vunv92/consumer/redis"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
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
