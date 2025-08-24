package main

import (
	"context"
	"encoding/json"
	"github.com/brianvoe/gofakeit/v7"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	kafkago "github.com/segmentio/kafka-go"

	"github.com/oziev02/wb/internal/domain"
)

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func brokersEnv() []string {
	return strings.Split(env("KAFKA_BROKERS", "localhost:9092"), ",")
}

func fakeOrder() domain.Order {
	gofakeit.Seed(time.Now().UnixNano())
	uid := gofakeit.UUID()
	itemCount := rand.Intn(3) + 1

	items := make([]domain.Item, 0, itemCount)
	total := 0
	for i := 0; i < itemCount; i++ {
		price := rand.Intn(1000) + 100
		items = append(items, domain.Item{
			ChrtID:      gofakeit.Number(1000000, 9999999),
			TrackNumber: gofakeit.LetterN(12),
			Price:       price,
			RID:         gofakeit.UUID(),
			Name:        gofakeit.ProductName(),
			Sale:        rand.Intn(50),
			Size:        "M",
			TotalPrice:  price,
			NmID:        gofakeit.Number(100000, 999999),
			Brand:       gofakeit.Company(),
			Status:      202,
		})
		total += price
	}

	return domain.Order{
		OrderUID:    uid,
		TrackNumber: gofakeit.LetterN(12),
		Entry:       "WBIL",
		Delivery: domain.Delivery{
			Name:    gofakeit.Name(),
			Phone:   gofakeit.Phone(),
			Zip:     gofakeit.Zip(),
			City:    gofakeit.City(),
			Address: gofakeit.Address().Address,
			Region:  gofakeit.State(),
			Email:   gofakeit.Email(),
		},
		Payment: domain.Payment{
			Transaction:  uid,
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       total,
			PaymentDT:    time.Now().Unix(),
			Bank:         "alpha",
			DeliveryCost: rand.Intn(1000),
			GoodsTotal:   total,
			CustomFee:    0,
		},
		Items:           items,
		Locale:          "en",
		CustomerID:      gofakeit.Username(),
		DeliveryService: "meest",
		ShardKey:        strconv.Itoa(rand.Intn(10)),
		SmID:            rand.Intn(1000),
		DateCreated:     time.Now().UTC(),
		OofShard:        strconv.Itoa(rand.Intn(3)),
	}
}

func main() {
	topic := env("KAFKA_TOPIC", "orders")
	n := envInt("PRODUCE_N", 100)

	w := &kafkago.Writer{
		Addr:     kafkago.TCP(brokersEnv()...),
		Topic:    topic,
		Balancer: &kafkago.LeastBytes{},
	}
	defer func() {
		if err := w.Close(); err != nil {
			log.Printf("writer close: %v", err)
		}
	}()

	ctx := context.Background()
	for i := 0; i < n; i++ {
		o := fakeOrder()
		b, err := json.Marshal(o)
		if err != nil {
			log.Fatalf("marshal: %v", err)
		}
		if err := w.WriteMessages(ctx, kafkago.Message{Value: b}); err != nil {
			log.Fatalf("write: %v", err)
		}
		time.Sleep(time.Duration(rand.Intn(300)) * time.Millisecond)
	}
	log.Printf("produced %d messages to %s", n, topic)
}
