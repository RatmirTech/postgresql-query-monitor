package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/dreadew/go-common/pkg/logger"
	"github.com/hashicorp/vault/api"
	"github.com/joho/godotenv"
	"github.com/ratmirtech/postgresql-query-monitor/internal/collectors"
	"github.com/ratmirtech/postgresql-query-monitor/internal/config"
	_ "github.com/ratmirtech/postgresql-query-monitor/internal/models"
	client "github.com/ratmirtech/postgresql-query-monitor/internal/review"
	"github.com/ratmirtech/postgresql-query-monitor/internal/serverinfo"
	"github.com/ratmirtech/postgresql-query-monitor/internal/sqlfiles"
	"github.com/spf13/cobra"
)
var csiCmd = &cobra.Command{
	Use:   "csi",
	Short: "Collect server info, send for analysis: config and server info (not metrics)",
	Run: func(cmd *cobra.Command, args []string) {
		var cfg config.Config

		if err := cfg.Load(); err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		ctx := context.Background()

		vaultClient, err := api.NewClient(api.DefaultConfig())
		if err != nil {
			log.Fatalf("Failed to create Vault client: %v", err)
		}

		vaultClient.SetToken(cfg.VaultToken)

		vaultPath, err := cmd.Flags().GetString("vp")
		if err != nil {
			log.Fatalf("Failed to get vault path: %v", err)
		}
		if vaultPath == "" {
			log.Fatalf("Vault path is required")
		}

		isSchedulerTask, err := cmd.Flags().GetBool("st")
		if err != nil {
			log.Fatalf("Failed to get scheduler task flag: %v", err)
		}

		collector := serverinfo.NewServerInfoCollector(vaultClient, vaultPath)

		info, err := collector.CollectServerData(ctx)
		if err != nil {
			log.Fatalf("Failed to collect server data: %v", err)
		}

		info.Environment = cfg.Environment

		analyzerClient := client.NewClient(cfg.ReviewAPI.URL)

		log.Default().Println("✅ Collected server info, sending for analysis...")
		log.Default().Printf("Server info: %+v", info)

		report, err := analyzerClient.AnalyzeConfig(ctx, info, isSchedulerTask)
		if err != nil {
			log.Fatalf("Failed to analyze config: %v", err)
		}

		log.Default().Printf("✅ Received recommendation: %+v", report)
	},
}

var csfCmd = &cobra.Command{
	Use:   "csf",
	Short: "Collect SQL files and send for analysis",
	Run: func(cmd *cobra.Command, args []string) {
		// флаги
		dir, _ := cmd.Flags().GetString("dir")
		modeStr, _ := cmd.Flags().GetString("mode")
		migrationsPath, _ := cmd.Flags().GetString("vp")
		specificFiles, _ := cmd.Flags().GetStringSlice("files")
		ignoreFiles, _ := cmd.Flags().GetStringSlice("ignore")
		enableIgnore, _ := cmd.Flags().GetBool("enable-ignore")
		//reviewURL, _ := cmd.Flags().GetString("review-url")

		// режим
		var mode sqlfiles.SearchMode
		switch modeStr {
		case "migrations":
			mode = sqlfiles.MigrationsOnly
		case "specific":
			mode = sqlfiles.SpecificFiles
		default:
			mode = sqlfiles.AllSQLFiles
		}

		cfg := sqlfiles.SearchConfig{
			RootPath:          dir,
			Mode:              mode,
			MigrationsPath:    migrationsPath,
			SpecificFileNames: specificFiles,
			EnableIgnoreList:  enableIgnore,
			IgnoreFiles:       ignoreFiles,
		}

		files, err := sqlfiles.CollectSQLFiles(cfg)
		if err != nil {
			log.Fatalf("❌ Failed to collect SQL files: %v", err)
		}

		if len(files) == 0 {
			log.Println("ℹ️ No SQL files found")
			return
		}

		log.Printf("✅ Found %d SQL files", len(files))
		log.Println("Files:")
		for _, f := range files {
			log.Printf("- %s (migration: %v)", f.Path, f.IsMigration)
		}

		// Разделяем файлы
		var migrations []sqlfiles.SQLFile
		var normal []sqlfiles.SQLFile
		for _, f := range files {
			if f.IsMigration {
				migrations = append(migrations, f)
			} else {
				normal = append(normal, f)
			}
		}

		// // Создаём клиента
		// ctx := context.Background()
		// apiClient := client.NewClient(cfg.ReviewAPI.URL)

		// // Отправляем обычные файлы пачкой
		// if len(normal) > 0 {
		// 	var queries []models.QueryReviewRequest
		// 	for _, f := range normal {
		// 		queries = append(queries, models.QueryReviewRequest{
		// 			SQL:         f.Content,         // сам SQL
		// 			ThreadID:    f.Title,           // для идентификации файла
		// 			Environment: "production",      // можно брать из cfg
		// 		})
		// 	}
		// 	batch := models.BatchReviewRequest{
		// 		Queries:     queries,
		// 		Environment: "production",
		// 	}

		// 	resp, err := apiClient.ReviewBatchQueries(ctx, batch)
		// 	if err != nil {
		// 		log.Fatalf("❌ Failed to review batch queries: %v", err)
		// 	}
		// 	log.Printf("✅ Batch review response: %+v", resp)
		// }

		// // Миграции — по одной
		// for _, f := range migrations {
		// 	req := models.MigrationReviewRequest{
		// 		SQL:         f.Content,     // сам SQL
		// 		Environment: "production",  // можно брать из cfg
		// 	}
		// 	resp, err := apiClient.ReviewMigration(ctx, req)
		// 	if err != nil {
		// 		log.Fatalf("❌ Failed to review migration %s: %v", f.Title, err)
		// 	}
		// 	log.Printf("✅ Migration %s review response: %+v", f.Title, resp)
		// }

	},
}
var csmCmd = &cobra.Command{
	Use:   "csm",
	Short: "Collect server info and system metrics",
	Run: func(cmd *cobra.Command, args []string) {
		var cfg config.Config

		if err := cfg.Load(); err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		isSchedulerTask, err := cmd.Flags().GetBool("st")
		if err != nil {
			log.Fatalf("Failed to get scheduler task flag: %v", err)
		}

		collector := collectors.NewSysMetricsCollector()
		metrics := collector.Collect()

		ctx := context.Background()
		
		vaultClient, err := api.NewClient(api.DefaultConfig())
		if err != nil {
			log.Fatalf("Failed to create Vault client: %v", err)
		}
		
		vaultClient.SetToken(cfg.VaultToken)

		vaultPath, err := cmd.Flags().GetString("vp")
		if err != nil {
			log.Fatalf("Failed to get vault path: %v", err)
		}
		if vaultPath == "" {
			log.Fatalf("Vault path is required")
		}

		collectorSI := serverinfo.NewServerInfoCollector(vaultClient, vaultPath)
		info, err := collectorSI.CollectServerInfo(ctx)
		if err != nil {
			log.Fatalf("Failed to collect server info: %v", err)
		}
		
		analyzerClient := client.NewClient(cfg.ReviewAPI.URL)
		report, err := analyzerClient.AnalyzeSystemMetrics(ctx, metrics, info, cfg.Environment, isSchedulerTask)
		if err != nil {
			log.Fatalf("Failed to analyze system metrics: %v", err)
		}

		log.Printf("✅ Received recommendation: %+v", report)
	},
}

func init() {
	csiCmd.Flags().String("vp", "", "Vault path")
	csiCmd.Flags().Bool("st", false, "Is scheduler task")
	
	csfCmd.Flags().String("dir", ".", "Directory to scan")
	csfCmd.Flags().String("mode", "all", "Search mode: all | migrations | specific")
	csfCmd.Flags().String("vp", "", "Migrations path (used if --mode=migrations)")
	csfCmd.Flags().StringSlice("files", []string{}, "Specific file names (used if --mode=specific)")
	csfCmd.Flags().Bool("enable-ignore", false, "Enable ignore list")
	csfCmd.Flags().StringSlice("ignore", []string{}, "Files to ignore")

	csmCmd.Flags().String("vp", "", "Vault path")
	csmCmd.Flags().Bool("st", false, "Is scheduler task")

	rootCmd.AddCommand(csiCmd, csfCmd, csmCmd)
	
	rootCmd.AddCommand(csiCmd, csfCmd, csmCmd)
}
var rootCmd = &cobra.Command{
	Use:   "pgmon",
	Short: "PostgreSQL Query Monitor CLI",
	Long:  `A CLI tool for monitoring PostgreSQL queries.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	logger.Init()
	
	Execute()
}