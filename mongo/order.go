package mongo

import (
	"beeketing.com/beeketing-consumer-go"
	"github.com/spf13/viper"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// OrderService lorem
type OrderService struct {
	Session    *mgo.Session
	Collection *mgo.Collection
}

func NewOrderService(session *mgo.Session) *OrderService {
	s := &OrderService{Session: session}

	s.Collection = s.Session.DB(viper.GetString("mongodb.db")).C("Order")

	return s
}

// GetByID lorem
func (s *OrderService) GetByID(id int) (*consumer.Order, error) {
	var order *consumer.Order

	return order, nil
}

// CountByShopID lorem
func (s *OrderService) CountByShopID(shopID int) (int, error) {
	count, err := s.Collection.Find(bson.M{"shopId": shopID}).Count()

	return count, err
}

// CountByProductRefID lorem
func (s *OrderService) CountByProductRefID(shopID int, productRefID int) (int, error) {
	count, err := s.Collection.Find(bson.M{
		"shopId":                 shopID,
		"lineItems.productRefId": productRefID,
	}).Count()

	return count, err
}

func (s *OrderService) GetByCartToken(shopID int64, cartToken string) (*consumer.Order, error) {
	var order *consumer.Order

	err := s.Collection.Find(bson.M{
		"shopId":    shopID,
		"cartToken": cartToken,
	}).One(&order)

	return order, err
}
