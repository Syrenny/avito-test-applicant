package postgres

import (
	"context"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionManager struct {
	TRM *manager.Manager
}

func NewTransactionManager(pool *pgxpool.Pool) *TransactionManager {
	trm := manager.Must(trmpgx.NewDefaultFactory(pool))
	return &TransactionManager{TRM: trm}
}

// Do выполняет функцию в транзакции
func (m *TransactionManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.TRM.Do(ctx, fn)
}
