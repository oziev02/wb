package usecase

import (
	"github.com/oziev02/wb/internal/domain"
)

// интерфейс кэша для удобства моков и тестов.
type OrdersCachePort interface {
	Get(id string) (domain.Order, bool)
	Set(o domain.Order)
	BulkSet([]domain.Order)
}

type OrderService struct {
	repo  domain.OrderRepository
	cache OrdersCachePort
}

func NewOrderService(r domain.OrderRepository, c OrdersCachePort) *OrderService {
	return &OrderService{repo: r, cache: c}
}

func (s *OrderService) InitCache(limit int) error {
	orders, err := s.repo.LoadAll(limit)
	if err != nil {
		return err
	}
	s.cache.BulkSet(orders)
	return nil
}

func (s *OrderService) Ingest(o domain.Order) error {
	if err := o.Validate(); err != nil {
		return err
	}
	if err := s.repo.UpsertOrder(o); err != nil {
		return err
	}
	s.cache.Set(o)
	return nil
}

func (s *OrderService) Get(id string) (domain.Order, bool, error) {
	if o, ok := s.cache.Get(id); ok {
		return o, true, nil
	}
	o, ok, err := s.repo.GetByID(id)
	if err != nil || !ok {
		return domain.Order{}, false, err
	}
	s.cache.Set(o)
	return o, true, nil
}
