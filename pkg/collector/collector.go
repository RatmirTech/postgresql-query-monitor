package collector

import (
	"context"
	"fmt"
	"os"

	hvault "github.com/hashicorp/vault/api"
	"go.uber.org/zap"

	"github.com/dreadew/go-common/pkg/clients/db/impl"
	"github.com/dreadew/go-common/pkg/logger"
	"github.com/ratmirtech/postgresql-query-monitor/pkg/metrics"
	pm "github.com/ratmirtech/postgresql-query-monitor/pkg/prometheus"
	vstore "github.com/ratmirtech/postgresql-query-monitor/pkg/vault"
)

type Collector struct {
	promManager *pm.Manager
	vaultAddr   string
	vaultClient *hvault.Client
}

func NewCollector(vaultAddr string, promManager *pm.Manager) *Collector {
	// инициализация метрик и логгера
	metrics.InitDefaultMetrics(promManager)
	logger.Init()

	return &Collector{
		promManager: promManager,
		vaultAddr:   vaultAddr,
	}
}

// lazy-init Vault client
func (c *Collector) vault() (*hvault.Client, error) {
	if c.vaultClient != nil {
		return c.vaultClient, nil
	}
	cfg := hvault.DefaultConfig()
	cfg.Address = c.vaultAddr

	cl, err := hvault.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("vault client init failed: %w", err)
	}

	tok := os.Getenv("VAULT_TOKEN")
	if tok == "" {
		return nil, fmt.Errorf("VAULT_TOKEN is empty")
	}
	cl.SetToken(tok)

	c.vaultClient = cl
	return c.vaultClient, nil
}

// Collect собирает указанные метрики из базы и отправляет их в Prometheus
func (c *Collector) Collect(ctx context.Context, secretPath, dbName, host string, metricNames []string) error {
	vcl, err := c.vault()
	if err != nil {
		return err
	}

	cfg, err := vstore.GetConnectionConfig(ctx, vcl, secretPath)
	if err != nil {
		return fmt.Errorf("get DB config from Vault at %s failed: %w", secretPath, err)
	}

	clientWrap, err := impl.New(ctx, cfg)
	if err != nil {
		return fmt.Errorf("db client init failed: %w", err)
	}
	defer func() {
		if err := clientWrap.Close(); err != nil {
			logger.GetLogger().Warn("failed to close db client", zap.Error(err))
		}
	}()

	db := clientWrap.DB()

	// Сбор метрик
	for _, name := range metricNames {
		fn, ok := metrics.GetMetric(name)
		if !ok {
			return fmt.Errorf("metric %q is not registered", name)
		}

		val, extraLabels, err := fn(ctx, db)
		if err != nil {
			return fmt.Errorf("collect %q failed: %w", name, err)
		}

		labels := map[string]string{
			"db_name": dbName,
			"host":    host,
		}
		for k, v := range extraLabels {
			labels[k] = v
		}

		if err := c.promManager.SetGauge(name, labels, val); err != nil {
			return fmt.Errorf("prometheus SetGauge for %q failed: %w", name, err)
		}
	}

	return nil
}
