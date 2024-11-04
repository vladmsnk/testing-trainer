package transactor

import (
	"context"
	"errors"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const txKey = "tx"

type QueryEngine interface {
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type QueryEngineProvider interface {
	GetQueryEngine(ctx context.Context) QueryEngine
}

type TransactionManager struct {
	pool *pgxpool.Pool
}

func New(conn *pgxpool.Pool) (*TransactionManager, error) {
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
