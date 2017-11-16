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

	OrderService consumer.OrderService

	Collection *mgo.Collection
}

func NewProductService(session *mgo.Session, orderService consumer.OrderService) *ProductService {
	s := &ProductService{Session: session, OrderService: orderService}

	s.Collection = s.Session.DB(viper.GetString("mongodb.db")).C("Product")

	return s
}

// GetByID lorem
func (s *ProductService) GetByID(id int64) (*consumer.Product, error) {
	var product *consumer.Product

	err := s.Collection.Find(bson.M{"refId": id}).One(&product)

	return product, err
}

// GetByShopID lorem
func (s *ProductService) GetByShopID(id int) ([]*consumer.Product, error) {
	var products []*consumer.Product

	err := s.Collection.Find(bson.M{"shopId": id}).All(&products)

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
