package consumer

type Setting struct {
	ID        int64  `json:"id"`
	ShopID    int64  `json:"shop_id"`
	KeyString string `json:"key_string"`
	Value     string `json:"value"`
	AppCode   string `json:"app_code"`
	Type      string `json:"type"`
}

type SettingService interface {
	Get(shopID int64, appCode string, keyString string) (*Setting, error)
	GetByKeyStrings(shopID int64, appCode string, keyStrings []string) ([]*Setting, error)
}
