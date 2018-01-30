package consumer

type Cart struct {
	ID           uint64 `bson:"_id"`
	ContactRefID string `bson:"contactRefId"`
	CartToken    string `bson:"cartToken"`
	ShopID       int    `bson:"shopId"`
}

type CartService interface {
	GetAbandonedCarts(shopID int, updatedAtMin, updatedAtMax string) ([]*Cart, error)
	GetShopIDOfAbandonedCarts(updatedAtMin, updatedAtMax string) ([]*Cart, error)
	GetAbandonedCartsByCartTokens(cartTokens []string) ([]*Cart, error)
}
