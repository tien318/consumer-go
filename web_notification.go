package consumer

import (
	"time"
)

type WebNotification struct {
	ID           int64
	ShopID       int64
	ContactRefID string
	CartToken    string
	Subscription string
	Data         string
	Campaign     string
	Send         int
	SendAt       time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type WebNotificationService interface {
	Add(wn *WebNotification) (int64, error)
	GetNotificationToSend() ([]*WebNotification, error)
	UpdateSent(wn *WebNotification) error
}
