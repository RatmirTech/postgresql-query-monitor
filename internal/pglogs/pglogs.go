package pglogs

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/dreadew/go-common/pkg/clients/db"
	"github.com/dreadew/go-common/pkg/clients/db/impl"
	"github.com/dreadew/go-common/pkg/logger"
	"github.com/hashicorp/vault/api"
	"github.com/ratmirtech/postgresql-query-monitor/pkg/vault"
	"go.uber.org/zap"
)

// PGLogsCollector собирает реальные логи PostgreSQL из файлов через SQL
type PGLogsCollector struct {
	vaultClient *api.Client
	vaultPath   string
}

// NewPGLogsCollector создает новый коллектор
func NewPGLogsCollector(vaultClient *api.Client, vaultPath string) *PGLogsCollector {
	return &PGLogsCollector{
		vaultClient: vaultClient,
		vaultPath:   vaultPath,
	}
}

// Collect собирает логи PostgreSQL за последние logTimeSeconds
func (c *PGLogsCollector) Collect(ctx context.Context, logTimeSeconds int) (string, error) {
	cfg, err := vault.GetConnectionConfig(ctx, c.vaultClient, c.vaultPath)
	if err != nil {
		return "", fmt.Errorf("failed to get DB config from Vault: %w", err)
	}

	clientWrap, err := impl.New(ctx, cfg)
	if err != nil {
		return "", fmt.Errorf("db client init failed: %w", err)
	}
	defer func() {
		if err := clientWrap.Close(); err != nil {
			l := logger.GetLogger()
			if l != nil {
				l.Warn("failed to close db client", zap.Error(err))
			}
		}
	}()

	client := clientWrap.DB()

	// Получаем список лог-файлов
	files, err := c.listLogFiles(ctx, client)
	if err != nil {
		return "", fmt.Errorf("failed to list log files: %w", err)
	}

	// Фильтруем файлы по времени (оставляем только актуальные)
	cutoffTime := time.Now().Add(-time.Duration(logTimeSeconds) * time.Second)
	recentFiles := filterLogFilesByTime(files, cutoffTime)

	var allLogs []string

	// Читаем каждый файл
	for _, file := range recentFiles {
		content, err := c.readFile(ctx, client, "log/"+file)
		if err != nil {
			log.Printf("Warning: failed to read log file %s: %v", file, err)
			continue
		}

		// Фильтруем строки по времени
		filteredLines := filterLogLinesByTime(content, cutoffTime)
		if len(filteredLines) > 0 {
			allLogs = append(allLogs, filteredLines...)
		}
	}

	if len(allLogs) == 0 {
		return "No PostgreSQL logs found in the specified time window.", nil
	}

	// Сортируем по времени (опционально)
	sort.Strings(allLogs)

	return strings.Join(allLogs, "\n"), nil
}

// listLogFiles получает список файлов в директории логов
func (c *PGLogsCollector) listLogFiles(ctx context.Context, client db.DB) ([]string, error) {
	rows, err := client.QueryContext(ctx, db.Query{
		Name: "list_log_files",
		Raw:  "SELECT pg_ls_dir('log')",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list log directory: %w", err)
	}
	defer rows.Close()

	var files []string
	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return nil, err
		}
		if strings.HasSuffix(filename, ".log") || strings.HasSuffix(filename, ".csv") {
			files = append(files, filename)
		}
	}
	return files, nil
}

// filterLogFilesByTime фильтрует файлы по дате в имени
func filterLogFilesByTime(files []string, cutoff time.Time) []string {
	var recent []string
	layout := "2006-01-02_150405" // соответствует postgresql-%Y-%m-%d_%H%M%S.log

	for _, file := range files {
		// Извлекаем дату из имени файла
		// Пример: postgresql-2025-09-06_180000.log
		re := regexp.MustCompile(`postgresql-(\d{4}-\d{2}-\d{2}_\d{6})`)
		matches := re.FindStringSubmatch(file)
		if len(matches) < 2 {
			continue // не распознали — пропускаем
		}

		t, err := time.Parse(layout, matches[1])
		if err != nil {
			continue
		}

		if t.After(cutoff) || t.Equal(cutoff) {
			recent = append(recent, file)
		}
	}
	return recent
}

// readFile читает содержимое файла через pg_read_file
func (c *PGLogsCollector) readFile(ctx context.Context, client db.DB, path string) (string, error) {
	var content string
	err := client.QueryRowContext(ctx, db.Query{
		Name: "read_log_file",
		Raw:  "SELECT pg_read_file($1)",
	}, path).Scan(&content)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return content, nil
}

// filterLogLinesByTime фильтрует строки лога по времени (для CSV и stderr форматов)
func filterLogLinesByTime(content string, cutoff time.Time) []string {
	var lines []string
	layoutCSV := "2006-01-02 15:04:05.000 UTC" // для csvlog
	layoutStderr := "2006-01-02 15:04:05 UTC"  // для stderr

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Пытаемся распарсить время из начала строки
		var t time.Time
		var err error

		// Для CSV: "2025-09-06 18:00:00.123 UTC",...
		if strings.Contains(line, ",") {
			parts := strings.SplitN(line, ",", 2)
			if len(parts) > 0 {
				t, err = time.Parse(layoutCSV, parts[0])
			}
		} else {
			// Для stderr: "2025-09-06 18:00:00 UTC ..."
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				timePart := strings.Join(parts[:3], " ") // "2025-09-06 18:00:00 UTC"
				t, err = time.Parse(layoutStderr, timePart)
			}
		}

		if err == nil && t.After(cutoff) {
			lines = append(lines, line)
		}
	}

	return lines
}
