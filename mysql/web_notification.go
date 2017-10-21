package mysql

import (
	"database/sql"
	"time"

	"beeketing.com/beeketing-consumer-go"
)

type WebNotificationService struct {
	DB *sql.DB
}

func (s *WebNotificationService) Add(n *consumer.WebNotification) (int64, error) {
	n.CreatedAt = time.Now()
	n.UpdatedAt = time.Now()

	query := `
	INSERT INTO web_notifications (shop_id, subscription, contact_ref_id, cart_token, campaign, send_at, data, created_at, updated_at, send)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := s.DB.Exec(query, n.ShopID, n.Subscription, n.ContactRefID, n.CartToken, n.Campaign, n.SendAt, n.Data, n.CreatedAt, n.UpdatedAt, n.Send)

	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}
