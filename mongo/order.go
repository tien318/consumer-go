package mongo

import (
	"beeketing.com/beeketing-consumer-go"
	"github.com/spf13/viper"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// OrderService lorem
type OrderService struct {
	Session *mgo.Session
}

// GetByID lorem
func (s *OrderService) GetByID(id int) (*consumer.Order, error) {
	var order *consumer.Order

	return order, nil
}

// CountByShopID lorem
func (s *OrderService) CountByShopID(shopID int) (int, error) {
	c := s.Session.DB(viper.GetString("mongodb.db")).C("Order")

	count, err := c.Find(bson.M{"shopId": shopID}).Count()

	return count, err
}

// CountByProductRefID lorem
func (s *OrderService) CountByProductRefID(shopID int, productRefID int) (int, error) {
	c := s.Session.DB(viper.GetString("mongodb.db")).C("Order")

	count, err := c.Find(bson.M{
		"shopId":                 shopID,
		"lineItems.productRefId": productRefID,
	}).Count()

	return count, err
}
