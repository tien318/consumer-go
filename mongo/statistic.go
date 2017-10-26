package mongo

import (
	"time"

	consumer "beeketing.com/beeketing-consumer-go"
	"github.com/spf13/viper"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type StatisticService struct {
	Session    *mgo.Session
	Collection *mgo.Collection
}

func NewStatisticService(session *mgo.Session) *StatisticService {
	s := &StatisticService{Session: session}

	s.Collection = s.Session.DB(viper.GetString("mongodb.db")).C("Statistic")

	return s
}

func (s *StatisticService) Get(shopID int64, statisticType string, refID int64, timeType string, time time.Time) (*consumer.Statistic, error) {
	stat := new(consumer.Statistic)

	query := bson.M{
		"shopId":   shopID,
		"type":     statisticType,
		"refId":    refID,
		"timeType": timeType,
	}

	if timeType != "total" {
		query["time"] = time.Unix()
	}

	err := s.Collection.Find(query).One(&stat)

	return stat, err
}

func (s *StatisticService) Add(stat *consumer.Statistic) error {
	return s.Collection.Insert(stat)
}

func (s *StatisticService) Increase(stat *consumer.Statistic, key string, value int64) error {
	selector := bson.M{"_id": stat.ID}

	update := bson.M{
		"$inc": bson.M{
			"data." + key: value,
		},
	}

	err := s.Collection.Update(selector, update)

	return err
}
