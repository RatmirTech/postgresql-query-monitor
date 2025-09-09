package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/dreadew/go-common/pkg/logger"
	"github.com/hashicorp/vault/api"
	"github.com/joho/godotenv"
	client "github.com/ratmirtech/postgresql-query-monitor/internal/review"
	"github.com/ratmirtech/postgresql-query-monitor/internal/serverinfo"

	"github.com/ratmirtech/postgresql-query-monitor/internal/collectors"
	"github.com/ratmirtech/postgresql-query-monitor/internal/config"
	"github.com/ratmirtech/postgresql-query-monitor/internal/pglogs"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	logger.Init()

	var cfg config.Config

	if err := cfg.Load(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Создаем Vault клиент
	ctx := context.Background()
	vaultClient, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatalf("Failed to create Vault client: %v", err)
	}

	collector := serverinfo.NewServerInfoCollector(vaultClient, "db/app1")

	info, e := collector.CollectServerData(ctx)
	if e != nil {
		fmt.Errorf("")
	}
	info.Environment = "production"
	info.Config.EffectiveIOConcurrency = "4"

	log.Printf("✅ Collected server info: %+v", info)

	// Create client
	analyzerClient := client.NewClient("http://185.159.111.235:8000")

	// Analyze configuration
	report, err := analyzerClient.AnalyzeConfig(ctx, info, true)
	if err != nil {
		log.Fatalf("Failed to analyze config: %v", err)
	}

	// Print recommendation
	fmt.Printf("Recommendation:\n")
	fmt.Printf("Content: %s\n", report.Content)
	fmt.Printf("Criticality: %s\n", report.Criticality)
	fmt.Printf("Recommendation: %s\n", report.Recommendation)
}

/*
func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	logger.Init()

	var cfg config.Config
	var rootCmd = &cobra.Command{
		Use:   "pgmon",
		Short: "PostgreSQL Query Monitor CLI",
		Long:  `Collects system metrics, PostgreSQL logs, and performance data.`,
	}

	// === Команда collect ===
	collectCmd := &cobra.Command{
		Use:   "collect [sysmetrics|pglogs|serverinfo|sqlfiles]",
		Short: "Collect metrics or logs",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			if err := cfg.Load(); err != nil {
				log.Fatalf("Failed to load config: %v", err)
			}

			// Создаем Vault клиент
			vaultClient, err := api.NewClient(api.DefaultConfig())
			if err != nil {
				log.Fatalf("Failed to create Vault client: %v", err)
			}

			collectorType := strings.ToLower(args[0])
			var wg sync.WaitGroup

			switch collectorType {
			case "sysmetrics":
				wg.Add(1)
				go func() {
					defer wg.Done()
					outputFile, _ := cmd.Flags().GetString("output")
					if err := runSysMetricsCollector(ctx, &cfg, outputFile); err != nil {
						log.Printf("Error: %v", err)
					}
				}()

			case "pglogs":
				wg.Add(1)
				go func() {
					defer wg.Done()
					logTime, _ := cmd.Flags().GetInt("lgt")
					vaultPath, _ := cmd.Flags().GetString("vsrc")
					if vaultPath == "" {
						log.Fatal("--vsrc (Vault path) is required for pglogs")
					}
					outputFile, _ := cmd.Flags().GetString("output")
					if err := runPGLogsCollector(ctx, vaultClient, vaultPath, logTime, outputFile); err != nil {
						log.Printf("Error: %v", err)
					}
				}()
			case "serverinfo":
				wg.Add(1)
				go func() {
					defer wg.Done()
					vaultPath, _ := cmd.Flags().GetString("vsrc")
					if vaultPath == "" {
						log.Fatal("--vsrc (Vault path) is required for serverinfo")
					}
					outputFile, _ := cmd.Flags().GetString("output")
					if err := runServerInfoCollector(ctx, vaultClient, vaultPath, outputFile); err != nil {
						log.Printf("Error: %v", err)
					}
				}()

			case "sqlfiles":
				wg.Add(1)
				go func() {
					defer wg.Done()
					outputFile, _ := cmd.Flags().GetString("output")
					if err := runSQLFilesCollector(ctx, &cfg, outputFile); err != nil {
						log.Printf("Error: %v", err)
					}
				}()

			default:
				log.Fatalf("Unknown collector type: %s", args[0])
			}

			wg.Wait()
		},
	}

	// Флаги
	collectCmd.Flags().String("output", "", "Output file (optional, auto-generated if not set)")
	collectCmd.Flags().Int("lgt", 60, "Log time window in seconds (for pglogs)")
	collectCmd.Flags().String("vsrc", "", "Vault path for DB credentials (e.g., 'db/app1')")

	rootCmd.AddCommand(collectCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
*/
// --- Сбор системных метрик ---
func runSysMetricsCollector(ctx context.Context, cfg *config.Config, outputFile string) error {
	if outputFile == "" {
		outputFile = "sysmetrics_" + time.Now().Format("20060102_150405") + ".txt"
	}

	collector := collectors.NewSysMetricsCollector()
	metrics := collector.Collect()

	log.Printf("✅ Collected system metrics: CPU: %.1f%%, RAM: %d MB used", metrics.CPULoad, metrics.RAMUsed/1024/1024)

	return saveSysMetricsToFile(metrics, outputFile)
}

func runServerInfoCollector(ctx context.Context, vaultClient *api.Client, vaultPath string, outputFile string) error {
	if outputFile == "" {
		outputFile = "serverinfo_" + time.Now().Format("20060102_150405") + ".txt"
	}

	collector := serverinfo.NewServerInfoCollector(vaultClient, vaultPath)

	info, e := collector.CollectServerData(ctx)
	if e != nil {
		return fmt.Errorf("")
	}
	info.Environment = "production"

	log.Printf("✅ Collected server info: %+v", info)

	// Сохраняем info в одну строку JSON
	jsonStr, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal server info to JSON: %w", err)
	}
	return saveServerInfoToFile(string(jsonStr), outputFile)
}

func saveServerInfoToFile(info string, outputFile string) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	fmt.Fprintf(file, "PostgreSQL Server Information\n")
	fmt.Fprintf(file, "============================\n\n")
	fmt.Fprintf(file, "%s\n", info)

	log.Printf("💾 Server info saved to %s", outputFile)
	return nil
}

func saveSysMetricsToFile(metrics collectors.SystemMetrics, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	fmt.Fprintf(file, "System Metrics Report\n")
	fmt.Fprintf(file, "====================\n\n")
	fmt.Fprintf(file, "Timestamp: %s\n\n", metrics.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(file, "CPU Information:\n")
	fmt.Fprintf(file, "  Cores: %d\n", metrics.CPUCores)
	fmt.Fprintf(file, "  Load Ratio: %.1f%%\n\n", metrics.CPULoad)
	fmt.Fprintf(file, "Memory Information:\n")
	fmt.Fprintf(file, "  Total RAM: %d MB\n", metrics.RAMTotal/1024/1024)
	fmt.Fprintf(file, "  Used RAM: %d MB\n", metrics.RAMUsed/1024/1024)
	fmt.Fprintf(file, "  Free RAM: %d MB\n\n", metrics.RAMFree/1024/1024)
	fmt.Fprintf(file, "Disk Information:\n")
	fmt.Fprintf(file, "  Total Disk: %d GB\n", metrics.DiskTotal/1024/1024/1024)
	fmt.Fprintf(file, "  Used Disk: %d GB\n", metrics.DiskUsed/1024/1024/1024)
	fmt.Fprintf(file, "  Free Disk: %d GB\n\n", metrics.DiskFree/1024/1024/1024)
	fmt.Fprintf(file, "Go Runtime Information:\n")
	fmt.Fprintf(file, "  Goroutines: %d\n", metrics.Goroutines)
	fmt.Fprintf(file, "  GC Pauses: %d ns\n", metrics.GCPauses)
	fmt.Fprintf(file, "  Heap Alloc: %d MB\n", metrics.HeapAlloc/1024/1024)
	fmt.Fprintf(file, "  Heap Sys: %d MB\n", metrics.HeapSys/1024/1024)
	fmt.Fprintf(file, "  Stack In Use: %d KB\n", metrics.StackInUse/1024)

	log.Printf("💾 Metrics saved to %s", filename)
	return nil
}

// --- Сбор логов PostgreSQL ---
func runPGLogsCollector(ctx context.Context, vaultClient *api.Client, vaultPath string, logTime int, outputFile string) error {
	if outputFile == "" {
		outputFile = "pglogs_" + time.Now().Format("20060102_150405") + ".txt"
	}

	collector := pglogs.NewPGLogsCollector(vaultClient, vaultPath)
	logs, err := collector.Collect(ctx, logTime)
	if err != nil {
		return fmt.Errorf("failed to collect PG logs: %w", err)
	}

	log.Printf("✅ Collected PostgreSQL logs for last %d seconds", logTime)

	return savePGLogsToFile(logs, outputFile)
}

func savePGLogsToFile(logs string, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	fmt.Fprintf(file, "PostgreSQL Query Logs (Last %s)\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(file, "==================================\n\n")
	fmt.Fprint(file, logs)

	log.Printf("💾 Logs saved to %s", filename)
	return nil
}

func runSQLFilesCollector(ctx context.Context, cfg *config.Config, outputFile string) error {
	if outputFile == "" {
		outputFile = "sqlfiles_" + time.Now().Format("20060102_150405") + ".txt"
	}

	// Путь к SQL-файлам можно задать через окружение SQLFILES_PATH, иначе берём текущую директорию
	root := "."

	// Поведение по умолчанию (можно расширить при необходимости, если структура cfg изменится)
	ignoreMigrations := true
	enableIgnoreList := true
	var ignoreFiles = []string{
		"main.sql",
	}
	var files []struct {
		Title   string
		Content string
	}

	var walk func(string) error
	walk = func(dir string) error {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return err
		}
		for _, e := range entries {
			path := dir + "/" + e.Name()
			if e.IsDir() {
				// Пропускаем каталоги с миграциями при включённой опции
				if ignoreMigrations && (strings.Contains(strings.ToLower(path), "migrations") || strings.Contains(strings.ToLower(e.Name()), "migration")) {
					continue
				}
				if err := walk(path); err != nil {
					return err
				}
				continue
			}
			if strings.HasSuffix(strings.ToLower(e.Name()), ".sql") {
				// Игнор по списку (если включено)
				if enableIgnoreList {
					skip := false
					for _, name := range ignoreFiles {
						if name == e.Name() {
							skip = true
							break
						}
					}
					if skip {
						continue
					}
				}
				b, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				files = append(files, struct {
					Title   string
					Content string
				}{Title: e.Name(), Content: string(b)})
			}
		}
		return nil
	}

	if err := walk(root); err != nil {
		return fmt.Errorf("failed to collect SQL files: %w", err)
	}

	log.Printf("✅ Collected %d SQL files from %s", len(files), root)

	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	for _, f := range files {
		fmt.Fprintf(outFile, "----- %s -----\n\n", f.Title)
		fmt.Fprintln(outFile, f.Content)
		fmt.Fprintln(outFile, "\n\n")
	}

	log.Printf("💾 SQL files saved to %s", outputFile)
	return nil
}
