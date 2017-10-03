package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"beeketing.com/beeketing-consumer-go/config"

	_ "github.com/go-sql-driver/mysql"

	"github.com/spf13/viper"
)

var db *sql.DB

func init() {
	config.Load()
}

func main() {
	fmt.Println("sync contact")
	initMysql()
	defer db.Close()

	emails := getEmails(9561507)

	for _, email := range emails {
		_, err := http.PostForm("https://a.klaviyo.com/api/v1/list/LrnK26/members", url.Values{
			"api_key":       {"pk_56a51680ebf1a3d5cc93ba6a82ecfe7ecb"},
			"email":         {email},
			"confirm_optin": {"false"},
		})

		if err != nil {
			panic(err)
		}

		fmt.Println(email)
	}
}

func initMysql() {
	db, _ = sql.Open("mysql", viper.GetString("mysql.dsn"))

	db.SetConnMaxLifetime(time.Second * 20)
	db.SetMaxIdleConns(30)
	db.SetMaxOpenConns(30)

	err := db.Ping()
	if err != nil {
		log.Fatal(err)
	}
}

func getEmails(shopID int) []string {
	query := "select email from contacts where shop_id = ? and (tracker_type = 'subscriber' or is_subscribed_via_cbox = true);"
	rows, err := db.Query(query, shopID)
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

	return emails
}
