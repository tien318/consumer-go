package mysql

import (
	"database/sql"

	"beeketing.com/beeketing-consumer-go"
	log "github.com/sirupsen/logrus"
)

// AppShopService is todo
type AppShopService struct {
	DB *sql.DB
}

// GetByID is todo
func (s *AppShopService) GetByID(id int) (*consumer.AppShop, error) {
	appShop := &consumer.AppShop{}

	stmt, err := s.DB.Prepare("select id, app_id, shop_id, token_key from apps_shops where id = ?")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer stmt.Close()

	var tokenKey []byte

	err = stmt.QueryRow(id).Scan(&appShop.ID, &appShop.AppID, &appShop.ShopID, &tokenKey)

	appShop.TokenKey = string(tokenKey)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	return appShop, nil
}

// GetByAppID is todo
func (s *AppShopService) GetByAppID(appID int) ([]*consumer.AppShop, error) {
	ass := []*consumer.AppShop{}

	stmt, err := s.DB.Prepare("SELECT id, shop_id, token_key FROM apps_shops WHERE app_id = ?")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(appID)
	if err != nil {
		log.Println(err)
		return ass, err
	}
	defer rows.Close()

	for rows.Next() {
		var as = &consumer.AppShop{AppID: appID}
		var tokenKey []byte

		err := rows.Scan(&as.ID, &as.ShopID, &tokenKey)
		if err != nil {
			log.Fatal(err)
		}

		as.TokenKey = string(tokenKey)

		ass = append(ass, as)
	}

	err = rows.Err()

	if err != nil {
		log.Fatal(err)
	}

	return ass, nil
}
