package consumer

// Product lorem
type Product struct {
	ID             int
	RefID          int     `bson:"refId"`
	ImageSourceURL string  `bson:"imageSourceUrl"`
	Title          string  `bson:"title"`
	MinPrice       float64 `bson:"minPrice"`
}

// ProductService lorem
type ProductService interface {
	GetByID(id int64) (*Product, error)
	GetByShopID(shopID int) ([]*Product, error)
	GetDefaultStatisticsData(shopID int, refID int) []int
}
