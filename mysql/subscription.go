package mysql

import (
	"database/sql"

	"beeketing.com/beeketing-consumer-go"
)

// ShopService lorem
type SubscriptionService struct {
	DB *sql.DB
}

func (s *SubscriptionService) GetByCartToken(cartToken string) (*consumer.Subscription, error) {
	sub := &consumer.Subscription{}

	query := `
	SELECT id, shop_id, subscription, contact_ref_id, cart_token
	FROM web_push_subscriptions
	WHERE cart_token = ?`

	err := s.DB.QueryRow(query, cartToken).Scan(&sub.ID, &sub.ShopID, &sub.Subscription, &sub.ContactRefID, &sub.CartToken)

	if err != nil {
		return nil, err
	}

	return sub, nil
}
