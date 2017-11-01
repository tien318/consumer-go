package mongo

import (
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"

	"beeketing.com/beeketing-consumer-go"
	"github.com/spf13/viper"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// ProductService lorem
type ProductService struct {
	Session *mgo.Session

	OrderService *OrderService
}

// GetByID lorem
func (s *ProductService) GetByID(id int64) (*consumer.Product, error) {
	var product *consumer.Product

	c := s.Session.DB(viper.GetString("mongodb.db")).C("Product")

	err := c.Find(bson.M{"refId": id}).One(&product)

	return product, err
}

// GetByShopID lorem
func (s *ProductService) GetByShopID(id int) ([]*consumer.Product, error) {
	var products []*consumer.Product

	c := s.Session.DB(viper.GetString("mongodb.db")).C("Product")

	err := c.Find(bson.M{"shopId": id}).All(&products)

	return products, err
}

// GetDefaultStatisticsData lorem
func (s *ProductService) GetDefaultStatisticsData(shopID int, refID int) []int {
	var view, addToCart, purchase int = 0, 0, 0

	// query to mongo to get count order
	purchase, err := s.OrderService.CountByProductRefID(shopID, refID)

	if err != nil {
		log.Errorf("%s: %s", "Count Order by product ref_id failed", err)
		return []int{0, 0, 0}
	}

	rand.Seed(time.Now().UnixNano())
	addToCart = int(float32(purchase) * (rand.Float32() + 1))

	view = addToCart * (rand.Intn(10) + 10)

	return []int{view, addToCart, purchase}
}
