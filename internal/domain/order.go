package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"time"

	"github.com/go-playground/validator/v10"
)

type Order struct {
	OrderUID          string    `json:"order_uid"`
	TrackNumber       string    `json:"track_number"`
	Entry             string    `json:"entry"`
	Delivery          Delivery  `json:"delivery"`
	Payment           Payment   `json:"payment"`
	Items             []Item    `json:"items"`
	Locale            string    `json:"locale"`
	InternalSignature string    `json:"internal_signature"`
	CustomerID        string    `json:"customer_id"`
	DeliveryService   string    `json:"delivery_service"`
	ShardKey          string    `json:"shardkey"`
	SmID              int       `json:"sm_id"`
	DateCreated       time.Time `json:"date_created"`
	OofShard          string    `json:"oof_shard"`
}

type Delivery struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

type Payment struct {
	Transaction  string `json:"transaction"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDT    int64  `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

type Item struct {
	ChrtID      int    `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	RID         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmID        int    `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

var v = validator.New()

// структурная и содержательная валидация.
func (o *Order) Validate() error {
	if o.OrderUID == "" {
		return errors.New("order_uid is required")
	}
	if o.TrackNumber == "" {
		return errors.New("track_number is required")
	}
	if len(o.Items) == 0 {
		return errors.New("items must not be empty")
	}
	if o.Payment.Amount < 0 || o.Payment.GoodsTotal < 0 || o.Payment.DeliveryCost < 0 {
		return fmt.Errorf("amount fields must be >= 0")
	}
	if o.Delivery.Email != "" {
		if _, err := mail.ParseAddress(o.Delivery.Email); err != nil {
			return fmt.Errorf("invalid delivery.email")
		}
	}
	// опционально: ограничить Locale
	if o.Locale != "" && o.Locale != "ru" && o.Locale != "en" {
		return fmt.Errorf("unsupported locale")
	}
	return nil
}

func (o *Order) RawJSON() ([]byte, error) { return json.Marshal(o) }
