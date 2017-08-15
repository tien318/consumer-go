package main

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/nvvu/consumer/config"
	"github.com/nvvu/consumer/console"
	"github.com/nvvu/consumer/mysql"
	"github.com/nvvu/consumer/redis"
	log "github.com/sirupsen/logrus"
)

func init() {
	config.Load()

	redis.Init()
}

func main() {
	// Connect to database
	db, err := sql.Open("mysql", "root:root@/beeketing-platform")

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	// Create services
	ass := &mysql.AppShopService{DB: db}

	// Command
	cmd := &console.Command{Ass: ass}

	// cmd.Schedule()
	// for true {
	// }

	cmd.BuildJSONStatisticFile()
}
