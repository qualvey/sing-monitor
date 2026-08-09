package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"sing-monitor-server/api"
	"sing-monitor-server/collector"
	"sing-monitor-server/config"
	"sing-monitor-server/db"
	"sing-monitor-server/realtime"
)

var version = "dev"

func main() {
	configPath := flag.String("config", "/etc/sing-monitor/config.yaml", "config file path")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("sing-monitor %s\n", version)
		return
	}

	api.Version = version

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("[Main] Failed to load config: %v", err)
	}
	log.Printf("[Main] sing-monitor %s, config: %s", version, *configPath)

	// PostgreSQL
	if err := db.InitDB(cfg.DSN()); err != nil {
		log.Fatalf("[Main] PostgreSQL init failed: %v", err)
	}
	log.Printf("[Main] PostgreSQL connected (%s:%d/%s)", cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.DBName)

	// 实时推送器（WebSocket 数据源）
	rt := realtime.NewBroadcaster(cfg.RealTime.IntervalMS, cfg.RealTime.OnlineThresholdSec)
	rt.Start()
	defer rt.Stop()

	// 采集器
	go collector.StartCollector(cfg, rt)

	// REST API + WebSocket
	r := api.SetupRouter(cfg, rt)
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("[Main] API server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("[Main] Server failed: %v", err)
	}
	_ = time.Now
}
