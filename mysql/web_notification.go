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

func (s *WebNotificationService) GetNotificationToSend() ([]*consumer.WebNotification, error) {
	notifications := make([]*consumer.WebNotification, 0)
	now := time.Now()

	query := `
	SELECT id, shop_id, subscription, data
	FROM web_notifications
	WHERE send = 0 AND send_at <= ?`

	rows, err := s.DB.Query(query, now)

	if err != nil {
		return notifications, err
	}
	defer rows.Close()

	for rows.Next() {
		n := &consumer.WebNotification{}

		err := rows.Scan(&n.ID, &n.ShopID, &n.Subscription, &n.Data)

		if err != nil {
			return notifications, err
		}

		notifications = append(notifications, n)
	}

	if err = rows.Err(); err != nil {
		return notifications, err
	}

	return notifications, nil
}

func (s *WebNotificationService) UpdateSent(wn *consumer.WebNotification) error {
	query := `
	UPDATE web_notifications
	SET send = 1
	WHERE id = ?`

	_, err := s.DB.Exec(query, wn.ID)

	return err
}
