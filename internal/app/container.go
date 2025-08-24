package app

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oziev02/wb/internal/adapters/db/postgres"
	"github.com/oziev02/wb/internal/cache"
	"github.com/oziev02/wb/internal/usecase"
)

type Container struct {
	Cfg  Config
	Pool *pgxpool.Pool
	Svc  *usecase.OrderService
}

func NewContainer(ctx context.Context, cfg Config) (*Container, error) {
	pool, err := pgxpool.New(ctx, cfg.DBURL)
	if err != nil {
		return nil, fmt.Errorf("pgxpool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("db ping: %w", err)
	}

	repo := postgres.NewOrderRepo(pool)

	c := cache.NewOrdersCache(cfg.CacheCap, cfg.CacheTTL)
	svc := usecase.NewOrderService(repo, c)

	if err := svc.InitCache(cfg.CacheRestoreLimit); err != nil {
		return nil, fmt.Errorf("init cache: %w", err)
	}

	return &Container{Cfg: cfg, Pool: pool, Svc: svc}, nil
}
