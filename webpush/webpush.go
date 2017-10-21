package webpush

import (
	"bytes"
	"encoding/json"
	"log"

	webpush "github.com/sherclockholmes/webpush-go"
)

const (
	publicKey       = "BMtBecN46lJbB-97vhFUqVfYCA5x8GkepTnVLt699gOmIuiUlYXNEdlDjKOcuoUctvkldPH9fwKs6NGG53SaQAY"
	vapidPrivateKey = "pZjKhkkDZl19i-CBkPQlG1lmfwyvJEjSmUzGFO1nrkA"
)

func Send(subJSON, msg string) {

	// subJSON := `{"endpoint":"https://fcm.googleapis.com/fcm/send/doq7MpV4bOA:APA91bFfor-scOf9KhXBZZc3IGnO27cYMRbX_gfjoH_FbNUd8cogiOPy9K1o9ZelPfFt97FbausAUprNIx8ZxSWSAIpsMr-wKQsmwUWcQ9tOc84ZlmBnA24c8T9r5dJOFqZzcm9SfeIC","expirationTime":null,"keys":{"p256dh":"BEr7iytbtrOYmo3ulN5Kj7l_C4MdpDSzMQ3-38KSiY19N7lj8lDyVwd6pAZWQEK9lx66ahuhVqciy7Tvc1ST2yQ=","auth":"cNgOz9EKuoIbZdSgoTjcQg=="}}`

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

	if err != nil {
		log.Fatal(err)
	}
}
