package main

import (
	"bytes"
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

	emails := getEmails(9652158)

	for i, email := range emails {
		resp, err := http.PostForm("https://a.klaviyo.com/api/v1/list/NNy7hi/members", url.Values{
			"api_key":       {"pk_4a8b8fc1d4e9bdf81d2bf218c93a10ce2f"},
			"email":         {email},
			"confirm_optin": {"false"},
		})

		if err != nil {
			panic(err)
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)

		fmt.Println(i, email)
		fmt.Println(buf.String())
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
