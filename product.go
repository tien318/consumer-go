package consumer

// Product lorem
type Product struct {
	ID    int
	RefID int `bson:"refId"`
}

// ProductService lorem
type ProductService interface {
	GetByID(id int) (*Product, error)
	GetByShopID(shopID int) ([]*Product, error)
	GetDefaultStatisticsData(refID int) []int
}
