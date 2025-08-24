package domain

type OrderRepository interface {
	UpsertOrder(o Order) error
	GetByID(orderUID string) (Order, bool, error)
	LoadAll(limit int) ([]Order, error)
}
