package pusher

import "encoding/json"

type Response OtherResponse

func UnmarshalResponse(data []byte) (Response, error) {
	var r Response
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Response) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type OtherResponse struct {
	Checkouts []Checkout `json:"checkouts"`
}

type Checkout struct {
	AbandonedCheckoutUrl  string         `json:"abandoned_checkout_url"`
	BuyerAcceptsMarketing bool           `json:"buyer_accepts_marketing"`
	CartToken             string         `json:"cart_token"`
	ClosedAt              interface{}    `json:"closed_at"`
	CompletedAt           interface{}    `json:"completed_at"`
	CreatedAt             string         `json:"created_at"`
	Currency              string         `json:"currency"`
	Customer              Customer       `json:"customer"`
	CustomerLocale        string         `json:"customer_locale"`
	DeviceId              interface{}    `json:"device_id"`
	DiscountCodes         []interface{}  `json:"discount_codes"`
	Email                 string         `json:"email"`
	Gateway               interface{}    `json:"gateway"`
	Id                    int64          `json:"id"`
	LandingSite           string         `json:"landing_site"`
	LineItems             []LineItem     `json:"line_items"`
	LocationId            interface{}    `json:"location_id"`
	Name                  string         `json:"name"`
	Note                  interface{}    `json:"note"`
	NoteAttributes        []interface{}  `json:"note_attributes"`
	Phone                 interface{}    `json:"phone"`
	ReferringSite         string         `json:"referring_site"`
	ShippingAddress       Address        `json:"shipping_address"`
	ShippingLines         []ShippingLine `json:"shipping_lines"`
	Source                interface{}    `json:"source"`
	SourceIdentifier      interface{}    `json:"source_identifier"`
	SourceName            string         `json:"source_name"`
	SourceUrl             interface{}    `json:"source_url"`
	SubtotalPrice         string         `json:"subtotal_price"`
	TaxLines              []OtherTaxLine `json:"tax_lines"`
	TaxesIncluded         bool           `json:"taxes_included"`
	Token                 string         `json:"token"`
	TotalDiscounts        string         `json:"total_discounts"`
	TotalLineItemsPrice   string         `json:"total_line_items_price"`
	TotalPrice            string         `json:"total_price"`
	TotalTax              string         `json:"total_tax"`
	TotalWeight           int64          `json:"total_weight"`
	UpdatedAt             string         `json:"updated_at"`
	UserId                interface{}    `json:"user_id"`
}

type LineItem struct {
	AppliedDiscounts      []interface{} `json:"applied_discounts"`
	CompareAtPrice        interface{}   `json:"compare_at_price"`
	DestinationLocationId int64         `json:"destination_location_id"`
	FulfillmentService    string        `json:"fulfillment_service"`
	GiftCard              bool          `json:"gift_card"`
	Grams                 int64         `json:"grams"`
	Key                   string        `json:"key"`
	LinePrice             string        `json:"line_price"`
	OriginLocationId      int64         `json:"origin_location_id"`
	Price                 string        `json:"price"`
	ProductId             int64         `json:"product_id"`
	Properties            interface{}   `json:"properties"`
	Quantity              int64         `json:"quantity"`
	RequiresShipping      bool          `json:"requires_shipping"`
	Sku                   string        `json:"sku"`
	TaxLines              []TaxLine     `json:"tax_lines"`
	Taxable               bool          `json:"taxable"`
	Title                 string        `json:"title"`
	VariantId             int64         `json:"variant_id"`
	VariantTitle          string        `json:"variant_title"`
	Vendor                string        `json:"vendor"`
}

type TaxLine struct {
	CompareAt float64 `json:"compare_at"`
	Position  int64   `json:"position"`
	Price     string  `json:"price"`
	Rate      float64 `json:"rate"`
	Source    string  `json:"source"`
	Title     string  `json:"title"`
	Zone      string  `json:"zone"`
}

type Customer struct {
	AcceptsMarketing    bool        `json:"accepts_marketing"`
	CreatedAt           string      `json:"created_at"`
	DefaultAddress      Address     `json:"default_address"`
	Email               string      `json:"email"`
	FirstName           string      `json:"first_name"`
	Id                  int64       `json:"id"`
	LastName            string      `json:"last_name"`
	LastOrderId         interface{} `json:"last_order_id"`
	LastOrderName       interface{} `json:"last_order_name"`
	MultipassIdentifier interface{} `json:"multipass_identifier"`
	Note                interface{} `json:"note"`
	OrdersCount         int64       `json:"orders_count"`
	Phone               interface{} `json:"phone"`
	State               string      `json:"state"`
	Tags                string      `json:"tags"`
	TaxExempt           bool        `json:"tax_exempt"`
	TotalSpent          string      `json:"total_spent"`
	UpdatedAt           string      `json:"updated_at"`
	VerifiedEmail       bool        `json:"verified_email"`
}

type ShippingLine struct {
	ApiClientId                   interface{}   `json:"api_client_id"`
	AppliedDiscounts              []interface{} `json:"applied_discounts"`
	CarrierIdentifier             interface{}   `json:"carrier_identifier"`
	CarrierServiceId              interface{}   `json:"carrier_service_id"`
	Code                          string        `json:"code"`
	DeliveryCategory              interface{}   `json:"delivery_category"`
	Id                            string        `json:"id"`
	Markup                        string        `json:"markup"`
	Phone                         interface{}   `json:"phone"`
	Price                         string        `json:"price"`
	RequestedFulfillmentServiceId interface{}   `json:"requested_fulfillment_service_id"`
	Source                        string        `json:"source"`
	TaxLines                      []interface{} `json:"tax_lines"`
	Title                         string        `json:"title"`
	ValidationContext             interface{}   `json:"validation_context"`
}

type Address struct {
	Address1     string      `json:"address1"`
	Address2     string      `json:"address2"`
	City         string      `json:"city"`
	Company      interface{} `json:"company"`
	Country      string      `json:"country"`
	CountryCode  string      `json:"country_code"`
	CountryName  *string     `json:"country_name"` /* optional */
	CustomerId   *int64      `json:"customer_id"`  /* optional */
	Default      *bool       `json:"default"`      /* optional */
	FirstName    string      `json:"first_name"`
	Id           *int64      `json:"id"` /* optional */
	LastName     string      `json:"last_name"`
	Latitude     interface{} `json:"latitude"`
	Longitude    interface{} `json:"longitude"`
	Name         string      `json:"name"`
	Phone        interface{} `json:"phone"`
	Province     interface{} `json:"province"`
	ProvinceCode interface{} `json:"province_code"`
	Zip          string      `json:"zip"`
}

type OtherTaxLine struct {
	Price string  `json:"price"`
	Rate  float64 `json:"rate"`
	Title string  `json:"title"`
}
