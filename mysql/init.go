package mysql

import (
	"database/sql"
	"time"

	"github.com/spf13/viper"
)

func NewMysql() (*sql.DB, error) {
	db, err := sql.Open("mysql", viper.GetString("mysql.dsn"))
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(time.Second * 20)
	db.SetMaxIdleConns(30)
	db.SetMaxOpenConns(30)

	err = db.Ping()

	return db, err
}
