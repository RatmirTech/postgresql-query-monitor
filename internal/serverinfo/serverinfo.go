package serverinfo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dreadew/go-common/pkg/clients/db"
	"github.com/dreadew/go-common/pkg/clients/db/impl"
	"github.com/hashicorp/vault/api"
	"github.com/ratmirtech/postgresql-query-monitor/internal/models"
	"github.com/ratmirtech/postgresql-query-monitor/pkg/vault"
)

// ServerInfoCollector собирает информацию о сервере PostgreSQL
type ServerInfoCollector struct {
	vaultClient *api.Client
	vaultPath   string
}

// NewServerInfoCollector создает новый коллектор
func NewServerInfoCollector(vaultClient *api.Client, vaultPath string) *ServerInfoCollector {
	return &ServerInfoCollector{
		vaultClient: vaultClient,
		vaultPath:   vaultPath,
	}
}

func (c *ServerInfoCollector) CollectServerData(ctx context.Context) (models.ServerData, error) {
	var data models.ServerData

	config, err := c.CollectConfig(ctx)
	if err != nil {
		return data, fmt.Errorf("failed to collect config: %w", err)
	}
	data.Config = config

	serverInfo, err := c.CollectServerInfo(ctx)
	if err != nil {
		return data, fmt.Errorf("failed to collect server info: %w", err)
	}
	data.ServerInfo = serverInfo

	data.Environment = fmt.Sprintf("%s@%s/%s", data.ServerInfo.Version, data.ServerInfo.Host, data.ServerInfo.Database)

	return data, nil
}

func (c *ServerInfoCollector) CollectServerInfo(ctx context.Context) (models.ServerInfo, error) {
	clientWrap, err := CreateDbWrap(ctx, c)
	if err != nil {
		return models.ServerInfo{}, err
	}

	defer clientWrap.Close()

	client := clientWrap.DB()

	return GetServerInfo(ctx, client)
}

func CreateDbWrap(ctx context.Context, c *ServerInfoCollector) (db.DatabaseClient, error) {
	cfg, err := vault.GetConnectionConfig(ctx, c.vaultClient, c.vaultPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get DB config from Vault: %w", err)
	}

	clientWrap, err := impl.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("db client init failed: %w", err)
	}

	return clientWrap, nil
}

func GetServerInfo(ctx context.Context, client db.DB) (models.ServerInfo, error) {
	var info models.ServerInfo

	row := client.QueryRowContext(ctx, db.Query{
		Raw: `SELECT 
            version() as version,
            inet_server_addr() as host,
            current_database() as database`,
	})

	var raw struct {
		Version  string
		Host     sql.NullString
		Database string
	}

	if err := row.Scan(&raw.Version, &raw.Host, &raw.Database); err != nil {
		return info, fmt.Errorf("failed to scan server info: %w", err)
	}

	info.Version = raw.Version
	info.Database = raw.Database
	if raw.Host.Valid {
		info.Host = raw.Host.String
	} else {
		info.Host = "localhost"
	}

	return info, nil
}

func (c *ServerInfoCollector) CollectConfig(ctx context.Context) (models.Config, error) {
	clientWrap, err := CreateDbWrap(ctx, c)
	if err != nil {
		return models.Config{}, err
	}
	defer clientWrap.Close()

	return GetConfig(ctx, clientWrap.DB())
}

func GetConfig(ctx context.Context, client db.DB) (models.Config, error) {
	var config models.Config

	query := `
		SELECT 
			name, setting 
		FROM pg_settings 
		WHERE name IN (
			'shared_buffers',
			'effective_cache_size', 
			'maintenance_work_mem',
			'checkpoint_completion_target',
			'wal_buffers',
			'default_statistics_target',
			'random_page_cost',
			'effective_io_concurrency',
			'work_mem',
			'min_wal_size',
			'max_wal_size'
		)
	`

	rows, err := client.QueryContext(ctx, db.Query{Raw: query})
	if err != nil {
		return config, fmt.Errorf("failed to query config parameters: %w", err)
	}
	defer rows.Close()

	// Маппинг полей для установки значений
	fieldMap := map[string]*string{
		"shared_buffers":               &config.SharedBuffers,
		"effective_cache_size":         &config.EffectiveCacheSize,
		"maintenance_work_mem":         &config.MaintenanceWorkMem,
		"checkpoint_completion_target": &config.CheckpointCompletionTarget,
		"wal_buffers":                  &config.WalBuffers,
		"default_statistics_target":    &config.DefaultStatisticsTarget,
		"random_page_cost":             &config.RandomPageCost,
		"effective_io_concurrency":     &config.EffectiveIOConcurrency,
		"work_mem":                     &config.WorkMem,
		"min_wal_size":                 &config.MinWalSize,
		"max_wal_size":                 &config.MaxWalSize,
	}

	for rows.Next() {
		var name, setting string
		if err := rows.Scan(&name, &setting); err != nil {
			return config, fmt.Errorf("failed to scan config parameter: %w", err)
		}

		if field, exists := fieldMap[name]; exists {
			*field = setting
		}
	}

	if err := rows.Err(); err != nil {
		return config, fmt.Errorf("error iterating config rows: %w", err)
	}

	return config, nil
}
