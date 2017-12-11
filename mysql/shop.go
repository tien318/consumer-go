package mysql

import (
	"database/sql"
	"fmt"
	"strings"

	"beeketing.com/beeketing-consumer-go"
	log "github.com/sirupsen/logrus"
)

// ShopService lorem
type ShopService struct {
	DB *sql.DB
}

// GetByID lorem
func (s *ShopService) GetByID(id int) (*consumer.Shop, error) {
	shop := &consumer.Shop{}

	stmt, err := s.DB.Prepare("SELECT id, user_id, name, domain, public_domain, api_key, platform FROM shops where id = ?")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(id).Scan(&shop.ID, &shop.UserID, &shop.Name, &shop.Domain, &shop.PublicDomain, &shop.APIKey, &shop.Platform)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	return shop, nil
}

// GetByIDs lorem
func (s *ShopService) GetByIDs(ids []int) ([]*consumer.Shop, error) {
	shops := []*consumer.Shop{}

	strIDs := strings.Trim(strings.Join(strings.Split(fmt.Sprint(ids), " "), ","), "[]")

	stmt, err := s.DB.Prepare("SELECT id, user_id, name, domain, public_domain, api_key, platform FROM shops where id IN (" + strIDs + ")")
	if err != nil {
		log.Println(err)
		return shops, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()

	if err != nil {
		return shops, err
	}
	defer rows.Close()

	var name, domain, publicDomain, apiKey []byte
	for rows.Next() {
		shop := &consumer.Shop{}

		err := rows.Scan(&shop.ID, &shop.UserID, &name, &domain, &publicDomain, &apiKey, &shop.Platform)
		if err != nil {
			return shops, err
		}

		shop.Name = string(name)
		shop.Domain = string(domain)
		shop.PublicDomain = string(publicDomain)
		shop.APIKey = string(apiKey)

		shops = append(shops, shop)
	}

	if err = rows.Err(); err != nil {
		return shops, err
	}

	return shops, nil
}
