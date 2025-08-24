package usecase

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/oziev02/wb/internal/domain"
)

type repoMock struct {
	upsert func(o domain.Order) error
	get    func(id string) (domain.Order, bool, error)
	load   func(limit int) ([]domain.Order, error)
}

func (m repoMock) UpsertOrder(o domain.Order) error              { return m.upsert(o) }
func (m repoMock) GetByID(id string) (domain.Order, bool, error) { return m.get(id) }
func (m repoMock) LoadAll(limit int) ([]domain.Order, error)     { return m.load(limit) }

type cacheMock struct{ store map[string]domain.Order }

func (c *cacheMock) Get(id string) (domain.Order, bool) { o, ok := c.store[id]; return o, ok }
func (c *cacheMock) Set(o domain.Order)                 { c.store[o.OrderUID] = o }
func (c *cacheMock) BulkSet(arr []domain.Order) {
	for _, o := range arr {
		c.Set(o)
	}
}

func sample() domain.Order {
	return domain.Order{
		OrderUID: "u1", TrackNumber: "tn", Entry: "WBIL",
		Delivery:    domain.Delivery{Email: "a@b.co"},
		Payment:     domain.Payment{Transaction: "u1", Amount: 1, GoodsTotal: 1},
		Items:       []domain.Item{{Name: "x", Price: 1, TotalPrice: 1}},
		DateCreated: time.Now().UTC(),
	}
}

func TestIngest_Valid(t *testing.T) {
	r := repoMock{
		upsert: func(o domain.Order) error { return nil },
		load:   func(int) ([]domain.Order, error) { return nil, nil },
		get:    func(string) (domain.Order, bool, error) { return domain.Order{}, false, nil },
	}
	c := &cacheMock{store: map[string]domain.Order{}}
	s := NewOrderService(r, c)
	err := s.Ingest(sample())
	require.NoError(t, err)
	_, ok := c.store["u1"]
	require.True(t, ok)
}

func TestGet_FallbackToDB(t *testing.T) {
	o := sample()
	r := repoMock{
		upsert: func(o domain.Order) error { return nil },
		load:   func(int) ([]domain.Order, error) { return nil, nil },
		get:    func(string) (domain.Order, bool, error) { return o, true, nil },
	}
	c := &cacheMock{store: map[string]domain.Order{}}
	s := NewOrderService(r, c)
	got, ok, err := s.Get("u1")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, o.OrderUID, got.OrderUID)
}

func TestIngest_Invalid(t *testing.T) {
	r := repoMock{
		upsert: func(o domain.Order) error { return errors.New("should not be called") },
		load:   func(int) ([]domain.Order, error) { return nil, nil },
		get:    func(string) (domain.Order, bool, error) { return domain.Order{}, false, nil },
	}
	c := &cacheMock{store: map[string]domain.Order{}}
	s := NewOrderService(r, c)
	bad := sample()
	bad.Items = nil
	err := s.Ingest(bad)
	require.Error(t, err)
}
