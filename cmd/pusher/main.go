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

	consumer "beeketing.com/beeketing-consumer-go"
	"beeketing.com/beeketing-consumer-go/apps/pusher"
	"beeketing.com/beeketing-consumer-go/mongo"
	"beeketing.com/beeketing-consumer-go/redis"
	mgo "gopkg.in/mgo.v2"

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

	redis.Init()
}

var urlAbandonedProduct = ""

func main() {
	fmt.Println("Pusher consumer")

	db, err := mysql.NewMysql()
	defer db.Close()

	if err != nil {
		panic(err)
	}

	urlAbandonedProduct = viper.GetString("url_abandoned_product")

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

	result, err := redis.Client.HGetAll("pusher").Result()
	if err != nil {
		log.Errorf("Fail to get cart token to redis")
		return
	}

	if len(result) == 0 {
		return
	}

	// get shops
	shopIDs := make([]int, 0)
	cartTokens := make([]string, 0)

	for key, value := range result {
		// fields[0]: shopID
		// fields[1]: cartToken
		fields := strings.Split(key, ":")

		if len(fields) != 2 {
			log.Errorf("%s: %s | %s", "Invalid hash field", key, err)
			return
		}

		shopID, _ := strconv.Atoi(fields[0])
		shopIDs = append(shopIDs, shopID)
		cartTokens = append(cartTokens, value)
	}

	// Delete hash
	_, err = redis.Client.Del("pusher").Result()
	if err != nil {
		log.Errorf("Delete Cart token in redis failed, details: %v", err)
		return
	}

	session, err := mgo.Dial(viper.GetString("mongodb.url"))
	if err != nil {
		log.Fatalf("%s: %s", "Failed to connect to mongo", err)
	}

	productService = mongo.NewProductService(session, nil)
	cartService = mongo.NewCartService(session)

	shops, err := shopService.GetByIDs(shopIDs)
	if err != nil {
		log.Errorf("%s: %s", "Get Shops failed", err)
		return
	}

	countNotification := 0
	carts, err := cartService.GetAbandonedCartsByCartTokens(cartTokens)
	if err != nil {
		log.Errorf("%s: %s", "Get abandoned carts failed", err)
		return
	}

	// get * from web_push_subscriptions by cart_token
	for _, cart := range carts {
		for _, shop := range shops {
			if cart.ShopID == shop.ID {
				sub, err := subscriptionService.GetByCartToken(cart.CartToken)
				if err == sql.ErrNoRows {
					continue
				} else if err != nil {
					log.Errorf("%s: %s", "Get subscription failed", err)
					continue
				}

				settings, err := getSettings(shop)
				if err != nil {
					log.Errorf("%s: %s", "Get settings failed", err)
					return
				}

				if len(settings) == 0 {
					return
				}

				// get abandoned product from http://localhost:9000/browse_abandoned_products.json
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
						ShopID  int                 `json:"shop_id"`
						APIKey  string              `json:"api_key"`
					}{
						Title:   title,
						Body:    body,
						URL:     url,
						Icon:    icon,
						Actions: actions,
						ShopID:  shop.ID,
						APIKey:  shop.APIKey,
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
						log.Errorf("%s: %s", "Add web notification failed", err)
					} else {
						countNotification++
					}
				}
				if len(carts) > 0 || countNotification > 0 {
					log.Info("Shop: ", shop.Domain)
					log.Info("Count abandoned carts: ", len(carts))
					log.Info("Count Notifications: ", countNotification)
				}
			}
		}
	}

	// log.Info("Start Fetch Abandoned Checkout from Shopify")

	// session, err := mgo.Dial(viper.GetString("mongodb.url"))

	// if err != nil {
	// 	log.Fatalf("%s: %s", "Failed to connect to mongo", err)
	// }

	// productService = mongo.NewProductService(session, nil)
	// cartService = mongo.NewCartService(session)

	// // init time
	// updatedAtMin := time.Now().Add(-time.Minute * 20).Format(time.RFC3339)
	// updatedAtMax := time.Now().Add(-time.Minute * 15).Format(time.RFC3339)

	// // get appshops
	// appShops, err := appShopService.GetByAppID(app.ID)
	// if err != nil {
	// 	log.Errorf("%s: %s", "Get App Shops failed", err)
	// 	return
	// }

	// if len(appShops) == 0 {
	// 	return
	// }

	// // get shops
	// shopIds := make([]int, 0)
	// for _, appShop := range appShops {
	// 	shopIds = append(shopIds, appShop.ShopID)
	// }

	// shops, err := shopService.GetByIDs(shopIds)
	// if err != nil {
	// 	log.Errorf("%s: %s", "Get Shops failed", err)
	// 	return
	// }

	// log.Info("Total Shops: ", len(shops))
	// if len(shops) == 0 {
	// 	return
	// }

	// for _, appShop := range appShops {
	// 	for _, shop := range shops {
	// 		if shop.ID == appShop.ShopID {
	// 			go getAbandonedCarts(shop, updatedAtMin, updatedAtMax)
	// 			// if !shop.IsSupportAbandonedCheckout() || isTestShop(shop) {
	// 			// 	go getAbandonedCarts(shop, updatedAtMin, updatedAtMax)
	// 			// } else {
	// 			// 	go getAbandonedCheckouts(shop, appShop, updatedAtMin, updatedAtMax)
	// 			// }
	// 		}
	// 	}
	// }
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

// // Get abandoned carts
// func getAbandonedCarts(shop *consumer.Shop, updatedAtMin, updatedAtMax string) {
// 	// get setting by shop
// 	settings, err := getSettings(shop)
// 	if err != nil {
// 		log.Errorf("%s: %s", "Get settings failed", err)
// 		return
// 	}

// 	if len(settings) == 0 {
// 		return
// 	}

// 	countNotification := 0
// 	// get abandoned cart from mongo by shopID, success = false
// 	carts, err := cartService.GetAbandonedCarts(shop.ID, updatedAtMin, updatedAtMax)
// 	if err != nil {
// 		log.Errorf("%s: %s", "Get abandoned carts failed", err)
// 		return
// 	}

// 	// get * from web_push_subscriptions by cart_token
// 	for _, cart := range carts {
// 		sub, err := subscriptionService.GetByCartToken(cart.CartToken)
// 		if err == sql.ErrNoRows {
// 			continue
// 		} else if err != nil {
// 			log.Errorf("%s: %s", "Get subscription failed", err)
// 			continue
// 		}

// 		// get abandoned product from http://localhost:9000/browse_abandoned_products.json
// 		product, err := getAbandonedProduct(shop, sub)
// 		if err != nil {
// 			log.Errorf("%s: %s", "get Abandoned Product failed", err)
// 			continue
// 		}

// 		var icon = product.ImageSourceURL

// 		for _, setting := range settings {
// 			wn := &consumer.WebNotification{}

// 			wn.ShopID = int64(shop.ID)
// 			wn.CartToken = cart.CartToken
// 			wn.Subscription = sub.Subscription
// 			wn.ContactRefID = sub.ContactRefID
// 			wn.Send = 0
// 			wn.Campaign = "pusher_abandoned_checkout"

// 			title := strings.Replace(setting.Subject, "{store_name}", shop.Name, -1)
// 			body := strings.Replace(setting.Message, "{store_name}", shop.Name, -1)

// 			if product != nil {
// 				title = strings.Replace(title, "{item_name}", product.Title, -1)
// 				title = strings.Replace(title, "{price}", strconv.FormatFloat(product.MinPrice, 'f', 2, 32), -1)
// 				body = strings.Replace(body, "{item_name}", product.Title, -1)
// 				body = strings.Replace(body, "{price}", strconv.FormatFloat(product.MinPrice, 'f', 2, 32), -1)
// 			}

// 			url := shop.GetCartUrl()

// 			actions := make([]map[string]string, 0)
// 			for _, button := range setting.Buttons {
// 				action := make(map[string]string)
// 				action["title"] = button.Text
// 				action["action"] = button.Url

// 				actions = append(actions, action)
// 			}

// 			dataObj := struct {
// 				Title   string              `json:"title"`
// 				Body    string              `json:"body"`
// 				URL     string              `json:"url"`
// 				Icon    string              `json:"icon"`
// 				Actions []map[string]string `json:"actions"`
// 				ShopID  int                 `json:"shop_id"`
// 				APIKey  string              `json:"api_key"`
// 			}{
// 				Title:   title,
// 				Body:    body,
// 				URL:     url,
// 				Icon:    icon,
// 				Actions: actions,
// 				ShopID:  shop.ID,
// 				APIKey:  shop.APIKey,
// 			}

// 			data, _ := json.Marshal(dataObj)
// 			wn.Data = string(data)

// 			if setting.KeyString == "pusher_cart_reminder_15_mins" {
// 				wn.SendAt = time.Now()
// 			} else if setting.KeyString == "pusher_cart_reminder_1_hour" {
// 				wn.SendAt = time.Now().Add(time.Minute * 45)
// 			} else if setting.KeyString == "pusher_cart_reminder_24_hours" {
// 				wn.SendAt = time.Now().Add(time.Hour * 24)
// 			}

// 			_, err := webNotificationService.Add(wn)
// 			if err != nil {
// 				log.Errorf("%s: %s", "Add web notification failed", err)
// 			} else {
// 				countNotification++
// 			}
// 		}
// 	}

// 	if len(carts) > 0 || countNotification > 0 {
// 		log.Info("Shop: ", shop.Domain)
// 		log.Infof("Time: %s - %s", updatedAtMin, updatedAtMax)
// 		log.Info("Count abandoned carts: ", len(carts))
// 		log.Info("Count Notifications: ", countNotification)
// 	}
// }

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

// unused
func getAbandonedCheckouts(shop *consumer.Shop, appShop *consumer.AppShop, updatedAtMin, updatedAtMax string) {
	// setting
	settings, err := getSettings(shop)
	if err != nil {
		log.Errorf("%s: %s", "Get settings failed", err)
		return
	}

	if len(settings) == 0 {
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
		return reminderSettings, nil
	}

	for _, setting := range settings {
		r, err := pusher.UnmarshalReminderSetting([]byte(setting.Value))
		if err != nil {
			continue
		}

		if r.Enable {
			r.KeyString = setting.KeyString
			reminderSettings = append(reminderSettings, r)
		}
	}

	return reminderSettings, nil
}

// unused
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
