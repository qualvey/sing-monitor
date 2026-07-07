package collector

import (
	"context"
	"log"
	"strings"
	"time"
	"sing-monitor-server/db"
	"sing-monitor-server/models"

	"github.com/sagernet/sing-box/experimental/v2rayapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func StartCollector(grpcAddr string, interval time.Duration) {
	conn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to sing-box grpc: %v", err)
	}

	client := v2rayapi.NewStatsServiceClient(conn)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				collectStats(client)
			}
		}
	}()
}

func collectStats(client v2rayapi.StatsServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Collect System Stats
	sysResp, err := client.GetSysStats(ctx, &v2rayapi.SysStatsRequest{})
	if err != nil {
		log.Printf("Error getting sys stats: %v", err)
	} else {
		sysLog := models.SysStatLog{
			Goroutines: sysResp.NumGoroutine,
			AllocBytes: sysResp.Alloc,
			SysBytes:   sysResp.Sys,
			Uptime:     sysResp.Uptime,
			Timestamp:  time.Now(),
		}
		db.DB.Create(&sysLog)
	}

	// Collect User Traffic
	// Based on sing-box stats pattern, we query all patterns
	queryResp, err := client.QueryStats(ctx, &v2rayapi.QueryStatsRequest{
		Pattern: "", // Empty pattern gets all or use specific pattern if needed
		Reset_:   true, // Assuming sing-box reset is bool
	})
	if err != nil {
		log.Printf("Error querying user stats: %v", err)
		return
	}

	now := time.Now()
	userTraffic := make(map[string]*models.TrafficLog)

	for _, stat := range queryResp.Stat {
		// Example stat name: user>>>[email]>>>traffic>>>down
		parts := strings.Split(stat.Name, ">>>")
		if len(parts) >= 4 && parts[0] == "user" && parts[2] == "traffic" {
			tag := parts[1]
			direction := parts[3]

			if _, ok := userTraffic[tag]; !ok {
				userTraffic[tag] = &models.TrafficLog{
					Timestamp: now,
				}
			}

			if direction == "down" {
				userTraffic[tag].DownBytes = stat.Value
			} else if direction == "up" {
				userTraffic[tag].UpBytes = stat.Value
			}
		}
	}

	for tag, traffic := range userTraffic {
		// Ensure user exists
		var user models.User
		if err := db.DB.Where("tag = ?", tag).FirstOrCreate(&user, models.User{Tag: tag, CreatedAt: now}).Error; err != nil {
			log.Printf("Error ensuring user %s: %v", tag, err)
			continue
		}

		traffic.UserID = user.ID
		db.DB.Create(traffic)
	}
	log.Printf("Collected stats for %d users", len(userTraffic))
}
