package mongo

import (
	mgo "gopkg.in/mgo.v2"
	"beeketing.com/beeketing-consumer-go"
	"github.com/spf13/viper"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type CartService struct {
	Session *mgo.Session
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
		"shopId": shopID,
		"success": false,
		"updatedAt": bson.M{
			"$gte": dateMin,
			"$lte": dateMax,
		},
	}).All(&carts)

	return carts, err
}