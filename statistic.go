package consumer

import (
	"time"
)

type Statistic struct {
	ID       int64            `bson:"_id,omitempty"`
	ShopID   int64            `bson:"shopId,omitempty"`
	RefID    int64            `bson:"refId,omitempty"`
	Type     string           `bson:"type,omitempty"`
	Data     map[string]int64 `bson:"data,omitempty"`
	TimeType string           `bson:"timeType,omitempty"`
	Time     int64            `bson:"time",omitempty`
}

type StatisticService interface {
	Get(shopID int64, statisticType string, refID int64, timeType string, time time.Time) (*Statistic, error)
	Add(s *Statistic) error
	Increase(stat *Statistic, key string, value int64) error
}
