package metrics

import (
	"context"
	"fmt"

	"github.com/dreadew/go-common/pkg/clients/db"
	"github.com/prometheus/client_golang/prometheus"
	pm "github.com/ratmirtech/postgresql-query-monitor/pkg/prometheus"
)

// MetricFunc описывает функцию сбора метрики
type MetricFunc func(ctx context.Context, client db.DB) (float64, prometheus.Labels, error)

// registry хранит все зарегистрированные функции метрик
var (
	registry          = map[string]MetricFunc{}
	systemMetricNames = []string{}
)

// RegisterMetric регистрирует функцию метрики
func RegisterMetric(name string, fn MetricFunc) error {
	if _, exists := registry[name]; exists {
		return fmt.Errorf("metric %s already registered", name)
	}
	registry[name] = fn
	return nil
}

// GetMetric возвращает функцию метрики по имени
func GetMetric(name string) (MetricFunc, bool) {
	fn, ok := registry[name]
	return fn, ok
}

// InitDefaultMetrics регистрирует базовые метрики и их функции
func InitDefaultMetrics(manager *pm.Manager) {
	// Gauge: количество активных соединений
	manager.RegisterGauge("db_active_connections", "Количество активных соединений PostgreSQL", []string{"db_name", "host"})
	RegisterMetric("db_active_connections", func(ctx context.Context, client db.DB) (float64, prometheus.Labels, error) {
		var cnt int
		q := db.Query{Raw: "SELECT count(*) FROM pg_stat_activity"}
		row := client.QueryRowContext(ctx, q)
		if err := row.Scan(&cnt); err != nil {
			return 0, nil, err
		}
		// Пример: значения лейблов должны быть заданы явно
		labels := prometheus.Labels{"db_name": "app1", "host": "localhost"}
		return float64(cnt), labels, nil
	})

	// Gauge: версия PostgreSQL
	manager.RegisterGauge("db_pg_version", "Версия сервера PostgreSQL", []string{"db_name", "host", "version"})
	RegisterMetric("db_pg_version", func(ctx context.Context, client db.DB) (float64, prometheus.Labels, error) {
		var version string
		q := db.Query{Raw: "SHOW server_version"}
		row := client.QueryRowContext(ctx, q)
		if err := row.Scan(&version); err != nil {
			return 0, nil, err
		}
		labels := prometheus.Labels{"db_name": "app1", "host": "localhost", "version": version}
		return 1, labels, nil
	})

}
