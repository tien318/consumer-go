package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"beeketing.com/beeketing-consumer-go/apps/pusher"

	"beeketing.com/beeketing-consumer-go"
	"beeketing.com/beeketing-consumer-go/config"
	"beeketing.com/beeketing-consumer-go/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
)

const (
	appCode = "pusher"
)

var (
	app                    *consumer.App
	appService             consumer.AppService
	appShopService         consumer.AppShopService
	shopService            consumer.ShopService
	subscriptionService    consumer.SubscriptionService
	settingService         consumer.SettingService
	webNotificationService consumer.WebNotificationService
	keyStrings             = []string{
		"pusher_cart_reminder_15_mins",
		"pusher_cart_reminder_1_hour",
		"pusher_cart_reminder_24_hours",
	}
)

func init() {
	config.Load()
}

func main() {
	fmt.Println("Pusher consumer")

	db, err := mysql.NewMysql()
	defer db.Close()

	if err != nil {
		panic(err)
	}

	appService = &mysql.AppService{DB: db}
	appShopService = &mysql.AppShopService{DB: db}
	shopService = &mysql.ShopService{DB: db}
	subscriptionService = &mysql.SubscriptionService{DB: db}
	settingService = &mysql.SettingService{DB: db}
	webNotificationService = &mysql.WebNotificationService{DB: db}

	app, err = appService.GetByAppCode(appCode)
	if err != nil && err != sql.ErrNoRows {
		log.Errorf("%s: %s", "Get App failed", err)
		return
	}

	if err == sql.ErrNoRows {
		log.Info("App Pusher NOT FOUND")
		return
	}

	// run()

	c := cron.New()

	log.Info("run every 5m")
	c.AddFunc("@every 5m", func() {
		run()
	})

	c.Start()

	select {}
}

func run() {
	log.Info("Start Fetch Abandoned Checkout from Shopify")

	// init time
	updatedAtMin := time.Now().Add(-time.Minute * 20).Format(time.RFC3339)
	updatedAtMax := time.Now().Add(-time.Minute * 15).Format(time.RFC3339)
	// log.Infof("Time: %s - %s", updatedAtMin, updatedAtMax)

	// get appshops
	appShops, err := appShopService.GetByAppID(app.ID)
	if err != nil {
		log.Errorf("%s: %s", "Get App Shops failed", err)
		return
	}

	// log.Info("Total AppShops: ", len(appShops))
	if len(appShops) == 0 {
		return
	}

	// get shops
	shopIds := make([]int, 0)
	for _, appShop := range appShops {
		shopIds = append(shopIds, appShop.ShopID)
	}

	shops, err := shopService.GetByIDs(shopIds)
	if err != nil {
		log.Errorf("%s: %s", "Get Shops failed", err)
	}

	log.Info("Total Shops: ", len(shops))
	if len(shops) == 0 {
		return
	}

	for _, appShop := range appShops {
		for _, shop := range shops {
			if shop.ID == appShop.ShopID {
				getAbandonedCheckouts(shop, appShop, updatedAtMin, updatedAtMax)
			}
		}
	}
}

func getAbandonedCheckouts(shop *consumer.Shop, appShop *consumer.AppShop, updatedAtMin, updatedAtMax string) {
	log.Info(">> Shop ID: ", shop.ID)

	// setting
	// log.Info("Get Setting")
	settings, err := getSettings(shop)
	if err != nil {
		log.Errorf("%s: %s", "Get settings failed", err)
		return
	}
	// log.Info("Count enable settings:", len(settings))

	// get checkout
	checkouts := fetchAbandonedCheckouts(shop, appShop, updatedAtMin, updatedAtMax)
	// log.Info("Count checkouts:", len(checkouts))

	countNotification := 0
	for _, checkout := range checkouts {
		sub, err := subscriptionService.GetByCartToken(checkout.CartToken)
		if err != nil && err != sql.ErrNoRows {
			log.Errorf("%s: %s", "get sub failed", err)
			continue
		}

		if err == sql.ErrNoRows {
			log.Info("No subscription found with cart token: ", checkout.CartToken)
			continue
		}

		// webpush.Send(sub.Subscription)
		for _, setting := range settings {
			wn := &consumer.WebNotification{}

			wn.ShopID = int64(shop.ID)
			wn.CartToken = checkout.CartToken
			wn.Subscription = sub.Subscription
			wn.ContactRefID = sub.ContactRefID
			wn.Send = 0
			wn.Campaign = "pusher_abandoned_checkout"

			title := strings.Replace(setting.Subject, "{store_name}", shop.Name, -1)
			body := strings.Replace(setting.Subject, "{store_name}", shop.Name, -1)
			url := "http://" + shop.Domain + "/cart?pusher=1"

			actions := make([]map[string]string, 0)
			for _, button := range setting.Buttons {
				action := make(map[string]string)
				action["title"] = button.Text
				action["action"] = button.Url

				actions = append(actions, action)
			}

			dataObj := struct {
				Title   string              `json:"title"`
				Body    string              `json:"body"`
				URL     string              `json:"url"`
				Actions []map[string]string `json:"actions"`
			}{
				Title:   title,
				Body:    body,
				URL:     url,
				Actions: actions,
			}

			data, _ := json.Marshal(dataObj)
			wn.Data = string(data)

			if setting.KeyString == "pusher_cart_reminder_15_mins" {
				wn.SendAt = time.Now().Add(time.Minute * 15)
			} else if setting.KeyString == "pusher_cart_reminder_1_hour" {
				wn.SendAt = time.Now().Add(time.Hour * 1)
			} else if setting.KeyString == "pusher_cart_reminder_24_hours" {
				wn.SendAt = time.Now().Add(time.Hour * 24)
			}

			_, err := webNotificationService.Add(wn)
			if err != nil {
				log.Errorf("%s: %s", "Error add web notification", err)
			} else {
				countNotification++
			}
		}
	}

	log.Info("Count Notifications:", countNotification)
}

func getSettings(shop *consumer.Shop) ([]pusher.ReminderSetting, error) {
	reminderSettings := make([]pusher.ReminderSetting, 0)
	settings, err := settingService.GetByKeyStrings(int64(shop.ID), appCode, keyStrings)

	if err != nil {
		return reminderSettings, err
	}

	if len(settings) == 0 {
		return reminderSettings, errors.New("No settings")
	}

	countEnable := 0
	for _, setting := range settings {
		r, err := pusher.UnmarshalReminderSetting([]byte(setting.Value))
		if err != nil {
			continue
		}

		if r.Enable {
			r.KeyString = setting.KeyString
			reminderSettings = append(reminderSettings, r)
			countEnable++
		}
	}

	if countEnable == 0 {
		return reminderSettings, errors.New("No enabled setting")
	}

	return reminderSettings, nil
}

func fetchAbandonedCheckouts(shop *consumer.Shop, appShop *consumer.AppShop, updatedAtMin, updatedAtMax string) []pusher.Checkout {
	checkouts := make([]pusher.Checkout, 0)

	for page := 1; ; page++ {
		// call api
		url := "https://" + shop.Domain + "/admin/checkouts.json?access_token=" + appShop.TokenKey + "&updated_at_min=" + updatedAtMin + "&updated_at_max=" + updatedAtMax + "&limit=250&page=" + strconv.Itoa(page)
		log.Info(url)
		resp, err := http.Get(url)
		if err != nil {
			log.Errorf("%s: %s", url, err)
			break
		}

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Errorf("%s: %s", "read body failed", err)
			break
		}

		r, err := pusher.UnmarshalResponse(body)
		if err != nil {
			log.Errorf("%s: %s", "unmarshal response failed", err)
			break
		}

		if len(r.Checkouts) == 0 {
			break
		}

		checkouts = append(checkouts, r.Checkouts...)

		if len(r.Checkouts) < 250 {
			break
		}
	}

	return checkouts
}
