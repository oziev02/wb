package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	kafkago "github.com/segmentio/kafka-go"

	"github.com/oziev02/wb/internal/domain"
	"github.com/oziev02/wb/internal/usecase"
)

type Config struct {
	Brokers []string
	Topic   string
	GroupID string
}

type Consumer struct {
	reader *kafkago.Reader
	uc     *usecase.OrderService
}

func NewConsumer(cfg Config, uc *usecase.OrderService) *Consumer {
	r := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:        cfg.Brokers,
		GroupID:        cfg.GroupID,
		Topic:          cfg.Topic,
		CommitInterval: time.Second,
	})
	return &Consumer{reader: r, uc: uc}
}

func (c *Consumer) Run(ctx context.Context) error {
	defer func() { _ = c.reader.Close() }()
	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			log.Printf("[kafka] fetch: %v", err)
			continue
		}
		var o domain.Order
		if err := json.Unmarshal(m.Value, &o); err != nil {
			log.Printf("[kafka] bad json at offset %d: %v", m.Offset, err)
			if err = c.reader.CommitMessages(ctx, m); err != nil {
				log.Printf("[kafka] commit after bad json: %v", err)
			}
			continue
		}
		if err := c.uc.Ingest(o); err != nil {
			log.Printf("[kafka] ingest failed, will retry (no commit): %v", err)
			continue
		}
		if err := c.reader.CommitMessages(ctx, m); err != nil {
			log.Printf("[kafka] commit: %v", err)
		}
	}
}
