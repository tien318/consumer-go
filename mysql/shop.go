package mysql

import (
	"database/sql"
	"fmt"
	"strings"

	"beeketing.com/consumer"
)

// ShopService lorem
type ShopService struct {
	DB *sql.DB
}

// GetByID lorem
func (s *ShopService) GetByID(id int) (*consumer.Shop, error) {
	shop := &consumer.Shop{}

	err := s.DB.QueryRow("SELECT id, user_id, name, domain, public_domain, api_key FROM shops where id = ?", id).Scan(&shop.ID, &shop.UserID, &shop.Name, &shop.Domain, &shop.PublicDomain, &shop.APIKey)

	if err != nil {
		return nil, err
	}

	return shop, nil
}

// GetByIDs lorem
func (s *ShopService) GetByIDs(ids []int) ([]*consumer.Shop, error) {
	shops := []*consumer.Shop{}

	strIDs := strings.Trim(strings.Join(strings.Split(fmt.Sprint(ids), " "), ","), "[]")

	rows, err := s.DB.Query("SELECT id, user_id, name, domain, public_domain, api_key FROM shops where id IN (?)", strIDs)

	if err != nil {
		return shops, err
	}
	defer rows.Close()

	for rows.Next() {
		shop := &consumer.Shop{}

		err := rows.Scan(&shop.ID, &shop.UserID, &shop.Name, &shop.Domain, &shop.PublicDomain, &shop.APIKey)
		if err != nil {
			return shops, err
		}

		shops = append(shops, shop)
	}

	return shops, nil
}
