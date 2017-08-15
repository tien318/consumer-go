package mysql

import (
	"database/sql"

	log "github.com/sirupsen/logrus"

	"bitbucket.org/vunv92/consumer"
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
		log.Fatal(err)
	}

	return shop, nil
}
