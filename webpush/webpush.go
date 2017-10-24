package webpush

import (
	"bytes"
	"encoding/json"
	"log"

	webpush "github.com/sherclockholmes/webpush-go"
)

const (
	publicKey       = "BOf3-fmrrg1U6kulLj0XF6O2YPTd7RfwTNXLIane5z2arxcTsAmajSKSgKmfBEFeMsWmR_kCBuAWH5btR6crMfE"
	vapidPrivateKey = "ZAtOP2Mir94xhOkCRCckWlw0s-nCHodOLyUVnPoXUm4"
)

func Send(subJSON, msg string) {
	// Decode subscription
	s := webpush.Subscription{}
	if err := json.NewDecoder(bytes.NewBufferString(subJSON)).Decode(&s); err != nil {
		log.Fatal(err)
	}

	// Send Notification
	_, err := webpush.SendNotification([]byte(msg), &s, &webpush.Options{
		Subscriber:      "<EMAIL@EXAMPLE.COM>",
		VAPIDPrivateKey: vapidPrivateKey,
	})

	// body, err := ioutil.ReadAll(resp.Body)
	// fmt.Println(string(body))

	if err != nil {
		log.Fatal(err)
	}
}
