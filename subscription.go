package consumer

import "time"

type Subscription struct {
	ID           int64     `json:"id"`
	ShopID       int64     `json:"shop_id"`
	ContactRefID string    `json:"contact_ref_id"`
	Subscription string    `json:"subscription"`
	CartToken    string    `json:"cart_token"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
}

type SubscriptionService interface {
	GetByCartToken(cartToken string) (*Subscription, error)
}
