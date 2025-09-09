package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/dreadew/go-common/pkg/logger"
	"github.com/hashicorp/vault/api"
	"github.com/joho/godotenv"
	"github.com/ratmirtech/postgresql-query-monitor/internal/config"
	client "github.com/ratmirtech/postgresql-query-monitor/internal/review"
	"github.com/ratmirtech/postgresql-query-monitor/internal/serverinfo"
	"github.com/spf13/cobra"
)
var csiCmd = &cobra.Command{
	Use:   "csi",
	Short: "Collect server info",
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

		_, err = analyzerClient.AnalyzeConfig(ctx, info, isSchedulerTask)
		if err != nil {
			log.Fatalf("Failed to analyze config: %v", err)
		}
	},
}

var csfCmd = &cobra.Command{
	Use:   "csf",
	Short: "Collect SQL files",
	Run: func(cmd *cobra.Command, args []string) {
		dir, _ := cmd.Flags().GetString("dir")
		pattern, _ := cmd.Flags().GetString("pattern")
		fmt.Printf("Collecting SQL files from %s with pattern %s\n", dir, pattern)
	},
}

var csmCmd = &cobra.Command{
	Use:   "csm",
	Short: "Collect system metrics",
	Run: func(cmd *cobra.Command, args []string) {
		interval, _ := cmd.Flags().GetInt("interval")
		output, _ := cmd.Flags().GetString("output")
		fmt.Printf("Collecting system metrics every %d seconds, output to %s\n", interval, output)
	},
}

func init() {
	csiCmd.Flags().String("vp", "", "Vault path")
	csiCmd.Flags().Bool("st", false, "Is scheduler task")

	csfCmd.Flags().String("dir", ".", "Directory to scan")
	csfCmd.Flags().String("pattern", "*.sql", "File pattern")

	csmCmd.Flags().Int("interval", 10, "Collection interval in seconds")
	csmCmd.Flags().String("output", "metrics.json", "Output file")

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