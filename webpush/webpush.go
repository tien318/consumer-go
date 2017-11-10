package webpush

import (
	"bytes"
	"encoding/json"

	log "github.com/sirupsen/logrus"

	webpush "github.com/sherclockholmes/webpush-go"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
	"github.com/sideshow/apns2/payload"
)

const (
	// chrome & firefox
	publicKey       = "BOf3-fmrrg1U6kulLj0XF6O2YPTd7RfwTNXLIane5z2arxcTsAmajSKSgKmfBEFeMsWmR_kCBuAWH5btR6crMfE"
	vapidPrivateKey = "ZAtOP2Mir94xhOkCRCckWlw0s-nCHodOLyUVnPoXUm4"
	// APNS
	certPath     = "../../Certificates.p12"
	certPassword = "bee@pusher@2017"
	webPushID    = "web.com.beeketing.pusher"
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

func SendApns(deviceToken, title, body, url string) {
	cert, err := certificate.FromP12File(certPath, certPassword)
	if err != nil {
		log.Error("Cert Error: ", err)
		return
	}

	notification := &apns2.Notification{}
	notification.DeviceToken = deviceToken
	notification.Topic = webPushID

	payload := payload.NewPayload().AlertTitle(title).AlertBody(body).AlertAction("View").URLArgs([]string{url})

	notification.Payload = payload

	client := apns2.NewClient(cert).Production()
	res, err := client.Push(notification)

	if err != nil {
		log.Error("Push APNs Error: ", err)
	}

	log.Infof("%v %v %v", res.StatusCode, res.ApnsID, res.Reason)
}
