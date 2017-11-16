package consumer

// Order lorem
type Order struct {
	ID int
}

// OrderService lorem
type OrderService interface {
	GetByID(id int) (*Order, error)
	CountByShopID(shopID int) (int, error)
	CountByProductRefID(shopID int, productID int) (int, error)
	GetByCartToken(cartToken string) (*Order, error)
}
