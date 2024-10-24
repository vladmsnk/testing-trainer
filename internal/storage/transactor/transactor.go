package transactor

import (
	"context"
	"errors"
	"fmt"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"testing_trainer/config"
)

const txKey = "tx"

type QueryEngine interface {
	Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
}

type QueryEngineProvider interface {
	GetQueryEngine(ctx context.Context) QueryEngine
}

type TransactionManager struct {
	pool *pgxpool.Pool
}

func New(cfg *config.Postgres) (*TransactionManager, error) {
	conn, err := pgxpool.New(context.Background(), cfg.GetConnectionString())
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}
	return &TransactionManager{pool: conn}, nil
}

func (tm *TransactionManager) RunRepeatableRead(ctx context.Context, fx func(ctxTX context.Context) error) error {
	tx, err := tm.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		return err
	}

	if err := fx(context.WithValue(ctx, txKey, tx)); err != nil {
		return errors.Join(err, tx.Rollback(ctx))
	}

	if err := tx.Commit(ctx); err != nil {
		return errors.Join(err, tx.Rollback(ctx))
	}
	return nil
}

func (tm *TransactionManager) GetQueryEngine(ctx context.Context) QueryEngine {
	tx, ok := ctx.Value(txKey).(QueryEngine)
	if ok && tx != nil {
		return tx
	}
	return tm.pool
}
