package mysql

import (
	"database/sql"
	"fmt"
	"strings"

	consumer "beeketing.com/beeketing-consumer-go"
)

type SettingService struct {
	DB *sql.DB
}

func (s *SettingService) Get(shopID int64, appCode string, keyString string) (*consumer.Setting, error) {
	setting := &consumer.Setting{}
	var value, sType []byte
	query := `
		SELECT id, value, type
		FROM settings
		WHERE shop_id = ? AND app_code = ? AND key_string = ?`

	err := s.DB.QueryRow(query, shopID, appCode, keyString).Scan(&setting.ID, &value, &sType)

	setting.KeyString = keyString
	setting.AppCode = appCode
	setting.Value = string(value)
	setting.Type = string(sType)

	if err != nil {
		return nil, err
	}

	return setting, nil
}

func (s *SettingService) GetByKeyStrings(shopID int64, appCode string, keyStrings []string) ([]*consumer.Setting, error) {
	settings := []*consumer.Setting{}
	strKeyStrings := "'" + strings.Trim(strings.Join(strings.Split(fmt.Sprint(keyStrings), " "), "','"), "[]") + "'"

	query := `
	SELECT id, value, type, key_string
	FROM settings
	WHERE shop_id = ? AND app_code = ? AND key_string IN (` + strKeyStrings + `)`

	rows, err := s.DB.Query(query, shopID, appCode)
	if err != nil {
		return settings, err
	}
	defer rows.Close()

	var value, sType []byte
	for rows.Next() {
		setting := &consumer.Setting{}

		err := rows.Scan(&setting.ID, &value, &sType, &setting.KeyString)
		if err != nil {
			return settings, err
		}

		setting.AppCode = appCode
		setting.Value = string(value)
		setting.Type = string(sType)

		settings = append(settings, setting)
	}

	if err = rows.Err(); err != nil {
		return settings, err
	}

	return settings, nil
}
