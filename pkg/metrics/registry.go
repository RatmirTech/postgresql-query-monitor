package metrics

import (
	"context"
	"fmt"

	"github.com/dreadew/go-common/pkg/clients/db"
	"github.com/prometheus/client_golang/prometheus"
	pm "github.com/ratmirtech/postgresql-query-monitor/pkg/prometheus"
)

type MetricFunc func(ctx context.Context, client db.DB) (float64, prometheus.Labels, error)

var registry = map[string]MetricFunc{}

func RegisterMetric(name string, fn MetricFunc) error {
	if _, exists := registry[name]; exists {
		return fmt.Errorf("metric %s already registered", name)
	}
	registry[name] = fn
	return nil
}

func GetMetric(name string) (MetricFunc, bool) {
	fn, ok := registry[name]
	return fn, ok
}

func InitDefaultMetrics(pmngr *pm.PrometheusManager) {
	pmngr.RegisterGauge("db_active_connections", "Количество активных соединений PostgreSQL", []string{"db_name", "host"})
	RegisterMetric("db_active_connections", func(ctx context.Context, client db.DB) (float64, prometheus.Labels, error) {
		var cnt int
		q := db.Query{Raw: "SELECT count(*) FROM pg_stat_activity"}
		row := client.QueryRowContext(ctx, q)
		if err := row.Scan(&cnt); err != nil {
			return 0, nil, err
		}
		return float64(cnt), nil, nil
	})

	pmngr.RegisterGauge("db_pg_version", "Версия сервера PostgreSQL", []string{"db_name", "host", "version"})
	RegisterMetric("db_pg_version", func(ctx context.Context, client db.DB) (float64, prometheus.Labels, error) {
		var version string
		q := db.Query{Raw: "SHOW server_version"}
		row := client.QueryRowContext(ctx, q)
		if err := row.Scan(&version); err != nil {
			return 0, nil, err
		}
		return 1, prometheus.Labels{"version": version}, nil
	})
}
