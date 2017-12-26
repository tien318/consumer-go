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
	"beeketing.com/beeketing-consumer-go/mongo"
	mgo "gopkg.in/mgo.v2"

	"beeketing.com/beeketing-consumer-go"
	"beeketing.com/beeketing-consumer-go/config"
	"beeketing.com/beeketing-consumer-go/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	//"bytes"
	"bytes"
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
	productService         consumer.ProductService
	cartService            consumer.CartService
	keyStrings             = []string{
		"pusher_cart_reminder_15_mins",
		"pusher_cart_reminder_1_hour",
		"pusher_cart_reminder_24_hours",
	}
)

func init() {
	config.Load()
}

var urlAbandonedProduct = ""

func main() {
	fmt.Println("Pusher consumer")

	db, err := mysql.NewMysql()
	defer db.Close()

	if err != nil {
		panic(err)
	}

	// session, err := mgo.Dial(viper.GetString("mongodb.url"))
	session, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{viper.GetString("mongodb.url")},
		Direct:   false,
		Timeout:  time.Second * 2,
		FailFast: true,
		Database: viper.GetString("mongodb.db"),
	})

	if err != nil {
		log.Fatalf("%s: %s", "Failed to connect to mongo", err)
	}
	defer session.Close()

	urlAbandonedProduct = viper.GetString("url_abandoned_product")

	appService = &mysql.AppService{DB: db}
	appShopService = &mysql.AppShopService{DB: db}
	shopService = &mysql.ShopService{DB: db}
	subscriptionService = &mysql.SubscriptionService{DB: db}
	settingService = &mysql.SettingService{DB: db}
	webNotificationService = &mysql.WebNotificationService{DB: db}
	productService = mongo.NewProductService(session, nil)
	cartService = mongo.NewCartService(session)

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
				if !shop.IsSupportAbandonedCheckout() || isTestShop(shop) {
					getAbandonedCarts(shop, updatedAtMin, updatedAtMax)
				} else {
					getAbandonedCheckouts(shop, appShop, updatedAtMin, updatedAtMax)
				}
			}
		}
	}
}

func isTestShop(shop *consumer.Shop) bool {
	testShopIds := []int{9555759, 9464465, 9565706, 9565217, 9569191, 9688252}
	for _, id := range testShopIds {
		if shop.ID == id {
			return true
		}
	}

	return false
}

// Get abandoned carts
func getAbandonedCarts(shop *consumer.Shop, updatedAtMin, updatedAtMax string) {
	log.Info(">> Shop domain: ", shop.Domain)
	log.Infof("Time: %s - %s", updatedAtMin, updatedAtMax)

	settings, err := getSettings(shop)
	if err != nil {
		log.Errorf("%s: %s", "Get settings failed", err)
		return
	}

	countNotification := 0
	carts, err := cartService.GetAbandonedCarts(shop.ID, updatedAtMin, updatedAtMax)
	if err != nil {
		log.Errorf("%s: %s", "Get abandoned carts failed", err)
		return
	}
	log.Info("Count abandoned carts:", len(carts))
	for _, cart := range carts {
		sub, err := subscriptionService.GetByCartToken(cart.CartToken)
		if err == sql.ErrNoRows {
			log.Info("No subscription found with cart token: ", cart.CartToken)
			continue
		} else if err != nil {
			log.Errorf("%s: %s", "Get subscription failed", err)
			continue
		}

		product, err := getAbandonedProduct(shop, sub)
		if err != nil {
			log.Errorf("%s: %s", "get Abandoned Product failed", err)
			continue
		}

		var icon = product.ImageSourceURL

		for _, setting := range settings {
			wn := &consumer.WebNotification{}

			wn.ShopID = int64(shop.ID)
			wn.CartToken = cart.CartToken
			wn.Subscription = sub.Subscription
			wn.ContactRefID = sub.ContactRefID
			wn.Send = 0
			wn.Campaign = "pusher_abandoned_checkout"

			title := strings.Replace(setting.Subject, "{store_name}", shop.Name, -1)
			body := strings.Replace(setting.Message, "{store_name}", shop.Name, -1)

			if product != nil {
				title = strings.Replace(title, "{item_name}", product.Title, -1)
				title = strings.Replace(title, "{price}", strconv.FormatFloat(product.MinPrice, 'f', 2, 32), -1)
				body = strings.Replace(body, "{item_name}", product.Title, -1)
				body = strings.Replace(body, "{price}", strconv.FormatFloat(product.MinPrice, 'f', 2, 32), -1)
			}

			url := shop.GetCartUrl()

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
				Icon    string              `json:"icon"`
				Actions []map[string]string `json:"actions"`
			}{
				Title:   title,
				Body:    body,
				URL:     url,
				Icon:    icon,
				Actions: actions,
			}

			data, _ := json.Marshal(dataObj)
			wn.Data = string(data)

			if setting.KeyString == "pusher_cart_reminder_15_mins" {
				wn.SendAt = time.Now()
			} else if setting.KeyString == "pusher_cart_reminder_1_hour" {
				wn.SendAt = time.Now().Add(time.Minute * 45)
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

func getAbandonedProduct(shop *consumer.Shop, sub *consumer.Subscription) (*consumer.Product, error) {
	params := struct {
		ShopID       string   `json:"shopId"`
		ContactRefID string   `json:"contactRefId"`
		BlackList    []string `json:"blackList"`
		Limit        int      `json:"limit"`
	}{
		ShopID: fmt.Sprintf("%d", shop.ID),
		// ShopID: fmt.Sprintf("%d", 9555759),
		ContactRefID: sub.ContactRefID,
		// ContactRefID: "9555759_1513137048883_1878",
		BlackList: []string{},
		Limit:     10,
	}

	jsonValues, _ := json.Marshal(params)
	log.Info("product ids: ", string(jsonValues))
	req, err := http.NewRequest("POST", urlAbandonedProduct, bytes.NewBuffer(jsonValues))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{
		Timeout: time.Duration(20 * time.Second),
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	json.Unmarshal(body, &data)

	productIds := data["items"].([]interface{})
	if len(productIds) == 0 {
		return nil, errors.New("No product found")
	}

	productIDStr, err := strconv.Atoi(productIds[0].(string))
	if err != nil {
		return nil, err
	}

	product, err := productService.GetByID(int64(productIDStr))
	// product, err := productService.GetByID(1)

	return product, err
}

func getAbandonedCheckouts(shop *consumer.Shop, appShop *consumer.AppShop, updatedAtMin, updatedAtMax string) {
	// log.Info(">> Shop ID: ", shop.ID)

	// setting
	settings, err := getSettings(shop)
	if err != nil {
		log.Errorf("%s: %s", "Get settings failed", err)
		return
	}

	// get checkout
	checkouts := fetchAbandonedCheckouts(shop, appShop, updatedAtMin, updatedAtMax)

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

		var icon = ""

		if len(checkout.LineItems) > 0 {
			product, err := productService.GetByID(checkout.LineItems[0].ProductId)
			if err == nil {
				icon = product.ImageSourceURL
				icon = strings.Replace(icon, ".jpg", "_small.jpg", -1)
				icon = strings.Replace(icon, ".png", "_small.png", -1)
			}
		}

		for _, setting := range settings {
			wn := &consumer.WebNotification{}

			wn.ShopID = int64(shop.ID)
			wn.CartToken = checkout.CartToken
			wn.Subscription = sub.Subscription
			wn.ContactRefID = sub.ContactRefID
			wn.Send = 0
			wn.Campaign = "pusher_abandoned_checkout"

			title := strings.Replace(setting.Subject, "{store_name}", shop.Name, -1)
			body := strings.Replace(setting.Message, "{store_name}", shop.Name, -1)

			if len(checkout.LineItems) > 0 {
				title = strings.Replace(title, "{item_name}", checkout.LineItems[0].Title, -1)
				title = strings.Replace(title, "{price}", checkout.LineItems[0].Price, -1)
				body = strings.Replace(body, "{item_name}", checkout.LineItems[0].Title, -1)
				body = strings.Replace(body, "{price}", checkout.LineItems[0].Price, -1)
			}

			url := "http://" + shop.Domain + "/cart"

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
				Icon    string              `json:"icon"`
				Actions []map[string]string `json:"actions"`
			}{
				Title:   title,
				Body:    body,
				URL:     url,
				Icon:    icon,
				Actions: actions,
			}

			data, _ := json.Marshal(dataObj)
			wn.Data = string(data)

			if setting.KeyString == "pusher_cart_reminder_15_mins" {
				wn.SendAt = time.Now()
			} else if setting.KeyString == "pusher_cart_reminder_1_hour" {
				wn.SendAt = time.Now().Add(time.Minute * 45)
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

	if countNotification > 0 {
		log.Info("Shop ", shop.ID, " : ", countNotification)
	}
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
		// log.Info(url)
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
