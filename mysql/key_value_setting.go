package mysql

import (
	"database/sql"
	"beeketing.com/beeketing-consumer-go"
	"fmt"
)

// KeyValueSettingService lorem
type KeyValueSettingService struct {
	DB *sql.DB
}

// GetByKeyName lorem
func (s *KeyValueSettingService) GetByKeyName(keyName string) (*consumer.KeyValueSetting, error) {
	keyValueSetting := &consumer.KeyValueSetting{}

	err := s.DB.QueryRow("select id, key_name, key_value from key_value_store where key_name = ?", keyName).Scan(&keyValueSetting.ID, &keyValueSetting.KeyName, &keyValueSetting.KeyValue)

	if err != nil {
		return nil, err
	}

	return keyValueSetting, nil
}

func (s *KeyValueSettingService) UpdateKeyValue(id int, keyValue string) (error) {
	query := fmt.Sprintf("update key_value_store set `key_value` = '%s' where id = %d", keyValue, id)
	rows, err := s.DB.Query(query)
	defer rows.Close()
	return err
}

func (s *KeyValueSettingService) CreateKeyValue(keyName string, keyValue string) (error) {
	query := fmt.Sprintf("insert into key_value_store (`key_name`, `key_value`) values ('%s', '%s')", keyName, keyValue)
	rows, err := s.DB.Query(query)
	defer rows.Close()
	return err
}