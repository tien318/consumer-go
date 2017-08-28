package mysql

import (
	"database/sql"

	"beeketing.com/consumer"
)

// AppService lorem
type AppService struct {
	DB *sql.DB
}

// GetByAppCode lorem
func (s *AppService) GetByAppCode(appCode string) (*consumer.App, error) {
	app := &consumer.App{}

	err := s.DB.QueryRow("select id, app_code, app_name from app where app_code = ?", appCode).Scan(&app.ID, &app.AppCode, &app.AppName)

	if err != nil {
		return nil, err
	}

	return app, nil
}
