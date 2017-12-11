package consumer

type Cart struct {
	ID           uint64 `bson:"_id"`
	ContactRefID string `bson:"contactRefId"`
	CartToken    string `bson:"cartToken"`
}

type CartService interface {
	GetAbandonedCarts(shopID int, updatedAtMin, updatedAtMax string) ([]*Cart, error)
}
