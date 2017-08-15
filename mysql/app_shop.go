package mysql

import (
	"database/sql"

	"github.com/nvvu/consumer"
	log "github.com/sirupsen/logrus"
)

// AppShopService is todo
type AppShopService struct {
	DB *sql.DB
}

// GetByID is todo
func (s *AppShopService) GetByID(id int) (*consumer.AppShop, error) {
	var appShop *consumer.AppShop

	return appShop, nil
}

// GetByAppID is todo
func (s *AppShopService) GetByAppID(appID int) ([]*consumer.AppShop, error) {
	ass := []*consumer.AppShop{}

	rows, err := s.DB.Query("SELECT id, shop_id FROM apps_shops WHERE app_id = ?", appID)

	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var as = &consumer.AppShop{AppID: appID}

		err := rows.Scan(&as.ID, &as.ShopID)
		if err != nil {
			log.Fatal(err)
		}

		ass = append(ass, as)
	}

	err = rows.Err()

	if err != nil {
		log.Fatal(err)
	}

	return ass, nil
}
