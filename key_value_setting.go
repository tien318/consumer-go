package consumer

// KeyValueSetting lorem
type KeyValueSetting struct {
	ID      int
	KeyName string
	KeyValue string
}

// KeyValueSettingService lorem
type KeyValueSettingService interface {
	GetByKeyName(keyName string) (*KeyValueSetting, error)
	UpdateKeyValue(id int, keyValue string) (error)
	CreateKeyValue(keyName string, keyValue string) (error)
}
