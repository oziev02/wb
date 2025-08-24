CREATE TABLE IF NOT EXISTS orders (
                                      order_uid TEXT PRIMARY KEY,
                                      track_number TEXT NOT NULL,
                                      entry TEXT NOT NULL,
                                      locale TEXT,
                                      internal_signature TEXT,
                                      customer_id TEXT,
                                      delivery_service TEXT,
                                      shardkey TEXT,
                                      sm_id INT,
                                      date_created TIMESTAMPTZ NOT NULL,
                                      oof_shard TEXT,
                                      raw_json JSONB NOT NULL
);

CREATE TABLE IF NOT EXISTS deliveries (
                                          order_uid TEXT PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
    name TEXT, phone TEXT, zip TEXT, city TEXT, address TEXT, region TEXT, email TEXT
    );

CREATE TABLE IF NOT EXISTS payments (
                                        order_uid TEXT PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
    transaction TEXT, request_id TEXT, currency TEXT, provider TEXT,
    amount INT, payment_dt BIGINT, bank TEXT, delivery_cost INT, goods_total INT, custom_fee INT
    );

CREATE TABLE IF NOT EXISTS items (
                                     id BIGSERIAL PRIMARY KEY,
                                     order_uid TEXT REFERENCES orders(order_uid) ON DELETE CASCADE,
    chrt_id INT, track_number TEXT, price INT, rid TEXT,
    name TEXT, sale INT, size TEXT, total_price INT, nm_id INT, brand TEXT, status INT
    );

CREATE INDEX IF NOT EXISTS idx_items_order_uid ON items(order_uid);
