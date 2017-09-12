package mongo

import (
	"log"
	"math/rand"
	"time"

	"beeketing.com/beeketing-consumer-go"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// ProductService lorem
type ProductService struct {
	Session *mgo.Session

	OrderService *OrderService
}

// GetByID lorem
func (s *ProductService) GetByID(id int) (*consumer.Product, error) {
	var product *consumer.Product

	return product, nil
}

// GetByShopID lorem
func (s *ProductService) GetByShopID(id int) ([]*consumer.Product, error) {
	var products []*consumer.Product

	c := s.Session.DB("beeketing-platform").C("Product")

	err := c.Find(bson.M{"shopId": id}).All(&products)

	return products, err
}

// GetDefaultStatisticsData lorem
func (s *ProductService) GetDefaultStatisticsData(refID int) []int {
	var view, addToCart, purchase int = 0, 0, 0

	// query to mongo to get count order
	purchase, err := s.OrderService.CountByProductRefID(refID)

	if err != nil {
		log.Fatal(err)
	}

	rand.Seed(time.Now().UnixNano())
	addToCart = int(float32(purchase) * (rand.Float32() + 1))

	view = addToCart * (rand.Intn(10) + 10)

	return []int{view, addToCart, purchase}
}
