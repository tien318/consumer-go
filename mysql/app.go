package mysql

import (
	"database/sql"

	"beeketing.com/beeketing-consumer-go"
	log "github.com/sirupsen/logrus"
)

// AppService lorem
type AppService struct {
	DB *sql.DB
}

// GetByAppCode lorem
func (s *AppService) GetByAppCode(appCode string) (*consumer.App, error) {
	app := &consumer.App{}
	stmt, err := s.DB.Prepare("select id, app_code, app_name from app where app_code = ?")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(appCode).Scan(&app.ID, &app.AppCode, &app.AppName)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	return app, nil
}
