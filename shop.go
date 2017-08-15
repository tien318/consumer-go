package consumer

// Shop is lorem
type Shop struct {
	ID   int
	Name string
}

// ShopService is lorem
type ShopService interface {
	GetById(id int) (*Shop, error)
	Add(s *Shop) error
	Update(s *Shop) error
	Delete(s *Shop) error
	DeleteById(id int) error
}
