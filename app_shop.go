package consumer

// AppShop is lorem
type AppShop struct {
	ID     int
	AppID  int
	ShopID int
}

// AppShopService is lorem
type AppShopService interface {
	GetByID(id int) (*AppShop, error)
	GetByAppID(appID int) ([]*AppShop, error)
}
