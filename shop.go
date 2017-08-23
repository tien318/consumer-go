package consumer

// Shop lorem
type Shop struct {
	ID           int
	UserID       int
	Name         string
	Domain       string
	PublicDomain string
	APIKey       string
}

// ShopService lorem
type ShopService interface {
	GetByID(id int) (*Shop, error)
	GetByIDs(ids []int) ([]*Shop, error)
}
