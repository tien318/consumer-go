package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

func main() {
	fmt.Println("Hello")

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
	query := "select email from contacts where shop_id = 9561507 and (tracker_type = 'subscriber' or is_subscribed_via_cbox = true);"
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	emails := make([]string, 0)

	for rows.Next() {
		var email string
		err := rows.Scan(&email)

		if err != nil {
			panic(err)
		}
		emails = append(emails, email)
	}

	fmt.Println(emails)
}
