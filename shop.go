package consumer

// Shop lorem
type Shop struct {
	ID           int
	UserID       int
	Name         string
	Domain       string
	PublicDomain string
	APIKey       string
	Platform     string
}

// Is shop support abandoned checkout api
func (s Shop) IsSupportAbandonedCheckout() bool {
	return s.Platform == "shopify"
}

// Get cart url
func (s Shop) GetCartUrl() string {
	var url = ""
	switch s.Platform {
	case "bigcommerce":
		url = "https://" + s.Domain + "/cart.php"
	default:
		url = "http://" + s.Domain + "/cart"
	}
	return url
}

// ShopService lorem
type ShopService interface {
	GetByID(id int) (*Shop, error)
	GetByIDs(ids []int) ([]*Shop, error)
}
