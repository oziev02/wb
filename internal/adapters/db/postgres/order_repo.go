package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oziev02/wb/internal/domain"
)

type OrderRepo struct {
	pool *pgxpool.Pool
}

func NewOrderRepo(pool *pgxpool.Pool) *OrderRepo { return &OrderRepo{pool: pool} }

func (r *OrderRepo) UpsertOrder(o domain.Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	raw, err := o.RawJSON()
	if err != nil {
		return fmt.Errorf("marshal raw: %w", err)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx) // безопасно: если уже commit — no-op
	}()

	_, err = tx.Exec(ctx, `
INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id,
                    delivery_service, shardkey, sm_id, date_created, oof_shard, raw_json)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
ON CONFLICT (order_uid) DO UPDATE SET
  track_number=EXCLUDED.track_number,
  entry=EXCLUDED.entry,
  locale=EXCLUDED.locale,
  internal_signature=EXCLUDED.internal_signature,
  customer_id=EXCLUDED.customer_id,
  delivery_service=EXCLUDED.delivery_service,
  shardkey=EXCLUDED.shardkey,
  sm_id=EXCLUDED.sm_id,
  date_created=EXCLUDED.date_created,
  oof_shard=EXCLUDED.oof_shard,
  raw_json=EXCLUDED.raw_json
`, o.OrderUID, o.TrackNumber, o.Entry, o.Locale, o.InternalSignature, o.CustomerID,
		o.DeliveryService, o.ShardKey, o.SmID, o.DateCreated, o.OofShard, raw)
	if err != nil {
		return fmt.Errorf("upsert orders: %w", err)
	}

	_, err = tx.Exec(ctx, `
INSERT INTO deliveries (order_uid, name, phone, zip, city, address, region, email)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
ON CONFLICT (order_uid) DO UPDATE SET
  name=EXCLUDED.name, phone=EXCLUDED.phone, zip=EXCLUDED.zip, city=EXCLUDED.city,
  address=EXCLUDED.address, region=EXCLUDED.region, email=EXCLUDED.email
`, o.OrderUID, o.Delivery.Name, o.Delivery.Phone, o.Delivery.Zip, o.Delivery.City,
		o.Delivery.Address, o.Delivery.Region, o.Delivery.Email)
	if err != nil {
		return fmt.Errorf("upsert deliveries: %w", err)
	}

	_, err = tx.Exec(ctx, `
INSERT INTO payments (order_uid, transaction, request_id, currency, provider,
                      amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
ON CONFLICT (order_uid) DO UPDATE SET
  transaction=EXCLUDED.transaction, request_id=EXCLUDED.request_id, currency=EXCLUDED.currency,
  provider=EXCLUDED.provider, amount=EXCLUDED.amount, payment_dt=EXCLUDED.payment_dt,
  bank=EXCLUDED.bank, delivery_cost=EXCLUDED.delivery_cost, goods_total=EXCLUDED.goods_total, custom_fee=EXCLUDED.custom_fee
`, o.OrderUID, o.Payment.Transaction, o.Payment.RequestID, o.Payment.Currency, o.Payment.Provider,
		o.Payment.Amount, o.Payment.PaymentDT, o.Payment.Bank, o.Payment.DeliveryCost, o.Payment.GoodsTotal, o.Payment.CustomFee)
	if err != nil {
		return fmt.Errorf("upsert payments: %w", err)
	}

	_, err = tx.Exec(ctx, `DELETE FROM items WHERE order_uid=$1`, o.OrderUID)
	if err != nil {
		return fmt.Errorf("delete items: %w", err)
	}

	for _, it := range o.Items {
		_, err = tx.Exec(ctx, `
INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
`, o.OrderUID, it.ChrtID, it.TrackNumber, it.Price, it.RID, it.Name, it.Sale, it.Size,
			it.TotalPrice, it.NmID, it.Brand, it.Status)
		if err != nil {
			return fmt.Errorf("insert item: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func (r *OrderRepo) GetByID(id string) (domain.Order, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var raw []byte
	err := r.pool.QueryRow(ctx, `SELECT raw_json FROM orders WHERE order_uid=$1`, id).Scan(&raw)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return domain.Order{}, false, err
		}
		return domain.Order{}, false, err
	}
	var o domain.Order
	if err := json.Unmarshal(raw, &o); err != nil {
		return domain.Order{}, false, fmt.Errorf("unmarshal: %w", err)
	}
	return o, true, nil
}

func (r *OrderRepo) LoadAll(limit int) ([]domain.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	rows, err := r.pool.Query(ctx, `SELECT raw_json FROM orders ORDER BY date_created DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Order
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		var o domain.Order
		if err := json.Unmarshal(raw, &o); err != nil {
			return nil, fmt.Errorf("unmarshal: %w", err)
		}
		out = append(out, o)
	}
	return out, rows.Err()
}
