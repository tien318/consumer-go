package mongo

import (
	"time"

	"beeketing.com/beeketing-consumer-go"
	"github.com/spf13/viper"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type CartService struct {
	Session    *mgo.Session
	Collection *mgo.Collection
}

func NewCartService(session *mgo.Session) *CartService {
	s := &CartService{Session: session}
	s.Collection = s.Session.DB(viper.GetString("mongodb.db")).C("Cart")
	return s
}

// get abandoned carts
func (s *CartService) GetAbandonedCarts(shopID int, updatedAtMin, updatedAtMax string) ([]*consumer.Cart, error) {
	var carts []*consumer.Cart

	dateMin, _ := time.Parse(time.RFC3339, updatedAtMin)
	dateMax, _ := time.Parse(time.RFC3339, updatedAtMax)
	err := s.Collection.Find(bson.M{
		"shopId":  shopID,
		"success": false,
		"updatedAt": bson.M{
			"$gte": dateMin,
			"$lte": dateMax,
		},
	}).All(&carts)

	return carts, err
}

// GetShopIDOfAbandonedCarts --
func (s *CartService) GetShopIDOfAbandonedCarts(updatedAtMin, updatedAtMax string) ([]*consumer.Cart, error) {
	var carts []*consumer.Cart

	dateMin, _ := time.Parse(time.RFC3339, updatedAtMin)
	dateMax, _ := time.Parse(time.RFC3339, updatedAtMax)
	err := s.Collection.Find(bson.M{
		"success": false,
		"updatedAt": bson.M{
			"$gte": dateMin,
			"$lte": dateMax,
		},
	}).All(&carts)

	return carts, err
}

// GetAbandonedCartsByCartTokens --
func (s *CartService) GetAbandonedCartsByCartTokens(cartTokens []string) ([]*consumer.Cart, error) {
	var carts []*consumer.Cart
	err := s.Collection.Find(bson.M{
		"success":   false,
		"cartToken": bson.M{"$in": cartTokens},
	}).All(&carts)

	return carts, err
}
