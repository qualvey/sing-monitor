package main

import (
	"fmt"
	"log"
	"time"

	"sing-monitor-server/api"
	"sing-monitor-server/collector"
	"sing-monitor-server/config"
	"sing-monitor-server/db"
)

func main() {
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize Database
	db.InitDB(cfg.DBPath)

	// Start Collector
	go collector.StartCollector(cfg.SingBoxGrpcAddr, time.Duration(cfg.CollectIntervalSeconds)*time.Second, cfg.DefaultCycleDays)

	// Start Gin Server
	r := api.SetupRouter()
	addr := fmt.Sprintf(":%d", cfg.APIServerPort)
	log.Printf("Starting sing-monitor API server on %s\n", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
