package impl

import (
	"context"
	"fmt"

	"github.com/dreadew/go-common/pkg/clients/db"
	pg_config "github.com/dreadew/go-common/pkg/config/pg"
	"github.com/dreadew/go-common/pkg/logger"

	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

type pgClient struct {
	masterDB db.DB
}

func New(ctx context.Context, config *pg_config.DbConfig) (db.DatabaseClient, error) {
	logger := logger.GetLogger()

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", config.Username, config.Password, config.Host, config.Port, config.Database, config.SSLMode)

	dbc, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		logger.Error("error while creating pg client", zap.String("error", err.Error()))
		return nil, err
	}

	return &pgClient{
		masterDB: &pg{dbc: dbc},
	}, nil
}

func (c *pgClient) DB() db.DB {
	return c.masterDB
}

func (c *pgClient) Close() error {
	if c.masterDB != nil {
		c.masterDB.Close()
	}

	return nil
}
