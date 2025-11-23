package postgres

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultMaxPoolSize = 1
)

type Postgres struct {
	Builder squirrel.StatementBuilderType
	Pool    *pgxpool.Pool
}

func New(url string, maxPool int) (*Postgres, error) {
	if maxPool <= 0 {
		maxPool = defaultMaxPoolSize
	}

	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("parse pgx config: %w", err)
	}

	config.MaxConns = int32(maxPool)

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("connect to pgx pool: %w", err)
	}

	return &Postgres{
		Pool:    pool,
		Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}, nil
}

func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}

func (p *Postgres) Ping(ctx context.Context) error {
	conn, err := p.Pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	return conn.Conn().Ping(ctx)
}
