package db

import (
	"context"

	"github.com/dreadew/go-common/pkg/interfaces"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

// Хэндлер для транзакции
type Handler func(ctx context.Context) error

// Интерфейс клиента к БД
type DatabaseClient interface {
	DB() DB
	Close() error
}

// Менеджер транзакций
type TxManager interface {
	ReadCommitted(ctx context.Context, f Handler) error
}

// Запрос к БД
type Query struct {
	Name string
	Raw  string
}

// Интерфейс клиента, поддерживающего
// выполнение запросов к БД
type SqlExecer interface {
	NamedExecer
	QueryExecer
}

// Интерфейс для работы с транзакциями
type Transactor interface {
	BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
}

// Интерфейс клиента, поддерживающего
// чтение данных из БД
type NamedExecer interface {
	ScanOneContext(ctx context.Context, dest interfaces.Entity, query Query, args ...interface{}) error
	ScanAllContext(ctx context.Context, dest interfaces.Entity, query Query, args ...interface{}) error
}

// Интерфейс клиента, поддерживающего
// выполнение запросов к БД
type QueryExecer interface {
	ExecContext(ctx context.Context, query Query, args ...interface{}) (pgconn.CommandTag, error)
	QueryContext(ctx context.Context, q Query, args ...interface{}) (pgx.Rows, error)
	QueryRowContext(ctx context.Context, q Query, args ...interface{}) pgx.Row
}

// Интерфейс для выполнения Ping к БД
type Pinger interface {
	Ping(ctx context.Context) error
}

// Интерфейс БД
type DB interface {
	SqlExecer
	Transactor
	Pinger
	Close()
}
