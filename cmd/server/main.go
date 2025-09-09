package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ratmirtech/postgresql-query-monitor/pkg/collector"
	pm "github.com/ratmirtech/postgresql-query-monitor/pkg/prometheus"
)

type CollectRequest struct {
	SecretPath  string   `json:"secret_path"`
	DBName      string   `json:"db_name"`
	Host        string   `json:"host"`
	MetricNames []string `json:"metric_names"`
}

func RunServer() {
	// Prometheus manager
	manager := pm.New()

	// Vault-адрес
	vaultAddr := "http://127.0.0.1:8200"
	coll := collector.NewCollector(vaultAddr, manager)

	mux := http.NewServeMux()

	// POST /collect
	mux.HandleFunc("/collect", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req CollectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
			return
		}

		ctx := context.Background()
		if err := coll.Collect(ctx, req.SecretPath, req.DBName, req.Host, req.MetricNames); err != nil {
			http.Error(w, fmt.Sprintf("collect failed: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("metrics collected"))
	})

	// GET /metrics
	mux.Handle("/metrics", pm.Handler())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Collector server listening on :%s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func main() {
	RunServer()
}
