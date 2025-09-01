package impl

import (
	"context"
	"log/slog"

	"github.com/dreadew/go-common/pkg/clients/db"
	db_constants "github.com/dreadew/go-common/pkg/constants/db"
	"github.com/dreadew/go-common/pkg/interfaces"
	"github.com/dreadew/go-common/pkg/logger"
	"github.com/dreadew/go-common/pkg/utils"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Тип для ключа транзакции
type key string

const (
	TxKey key = "tx"
)

type pg struct {
	dbc *pgxpool.Pool
}

func NewDB(dbc *pgxpool.Pool, logger *slog.Logger) db.DB {
	return &pg{
		dbc: dbc,
	}
}

func (p *pg) ScanOneContext(ctx context.Context, dest interfaces.Entity, q db.Query, args ...interface{}) error {
	p.logQuery(ctx, q, args...)
	row, err := p.QueryContext(ctx, q, args...)
	if err != nil {
		return err
	}

	return pgxscan.ScanOne(dest, row)
}

func (p *pg) ScanAllContext(ctx context.Context, dest interfaces.Entity, q db.Query, args ...interface{}) error {
	p.logQuery(ctx, q, args...)
	rows, err := p.QueryContext(ctx, q, args...)
	if err != nil {
		return err
	}

	return pgxscan.ScanAll(dest, rows)
}

func (p *pg) ExecContext(ctx context.Context, q db.Query, args ...interface{}) (pgconn.CommandTag, error) {
	p.logQuery(ctx, q, args...)
	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.Exec(ctx, q.Raw, args...)
	}

	return p.dbc.Exec(ctx, q.Raw, args...)
}

func (p *pg) QueryContext(ctx context.Context, q db.Query, args ...interface{}) (pgx.Rows, error) {
	p.logQuery(ctx, q, args...)
	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.Query(ctx, q.Raw, args...)
	}

	return p.dbc.Query(ctx, q.Raw, args...)
}

func (p *pg) QueryRowContext(ctx context.Context, q db.Query, args ...interface{}) pgx.Row {
	p.logQuery(ctx, q, args...)
	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.QueryRow(ctx, q.Raw, args...)
	}

	return p.dbc.QueryRow(ctx, q.Raw, args...)
}

func (p *pg) BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error) {
	return p.dbc.BeginTx(ctx, opts)
}

func (p *pg) Ping(ctx context.Context) error {
	return p.dbc.Ping(ctx)
}

func (p *pg) Close() {
	p.dbc.Close()
}

func MakeContextTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, TxKey, tx)
}

func (p *pg) logQuery(ctx context.Context, q db.Query, args ...interface{}) {
	logger := logger.GetLogger()
	query := utils.Pretty(q.Raw, db_constants.PlaceholderSignQuestion, args...)
	logger.Info(query)
}
