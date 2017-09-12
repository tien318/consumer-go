package statistic

import (
	"encoding/json"
	"time"
)

// ProductStat lorem
type ProductStat struct {
	Version int64            `json:"version"`
	Data    map[string][]int `json:"data"`
}

// NewProductStat create new ProductStat
func NewProductStat() *ProductStat {
	ps := &ProductStat{}
	ps.Version = time.Now().Unix()
	ps.Data = make(map[string][]int)
	return ps
}

// GetJSONData lorem
func (ps *ProductStat) GetJSONData() []byte {
	data, _ := json.Marshal(ps)

	return data
}
