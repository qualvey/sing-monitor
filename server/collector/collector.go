package collector

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"sing-monitor-server/config"
	"sing-monitor-server/db"
	"sing-monitor-server/models"
	"sing-monitor-server/realtime"

	"github.com/v2fly/v2ray-core/v5/app/stats/command"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm"
)

// StartCollector 启动采集循环：拉 sing-box gRPC 统计 → 增量入库 + 累计 upsert + 实时推送
func StartCollector(cfg *config.Config, rt *realtime.Broadcaster) {
	interval := cfg.PollIntervalDuration()
	conn, err := grpc.Dial(cfg.Singbox.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("[Collector] dial sing-box gRPC failed: %v", err)
		return
	}
	client := command.NewStatsServiceClient(conn)

	log.Printf("[Collector] Connected to sing-box gRPC at %s", cfg.Singbox.Address)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			collect(client, cfg, rt)
			// 动态间隔：有实时监控订阅者时切换到前端指定/默认高频，否则回默认
			target := interval
			if fast := rt.EffectivePollInterval(); fast > 0 {
				target = fast
			}
			if target != interval {
				ticker.Reset(target)
				interval = target
				log.Printf("[Collector] poll interval -> %s", interval)
			}
		}
	}
}

func collect(client command.StatsServiceClient, cfg *config.Config, rt *realtime.Broadcaster) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 分 user / inbound 两类查询，避免大响应
	categories := []struct{ pattern, category string }{
		{"user>>>", "user"},
		{"inbound>>>", "inbound"},
	}
	for _, c := range categories {
		queryPattern := c.pattern
		if cfg.Singbox.Pattern != "" {
			queryPattern = cfg.Singbox.Pattern + queryPattern
		}
		resp, err := client.QueryStats(ctx, &command.QueryStatsRequest{
			Pattern: queryPattern,
			Reset_:  cfg.Singbox.ResetOnQuery,
		})
		if err != nil {
			log.Printf("[Collector] QueryStats(%s) error: %v", c.category, err)
			continue
		}
		now := time.Now()
		perTarget := make(map[string]*[2]int64) // tag -> [up, down]

		for _, stat := range resp.Stat {
			// 名称格式: user>>><tag>>>>traffic>>>uplink|downlink 或 inbound>>>...
			parts := strings.Split(stat.Name, ">>>")
			if len(parts) < 4 || parts[0] != c.category || parts[2] != "traffic" {
				continue
			}
			tag := parts[1]
			dir := parts[3]
			entry, ok := perTarget[tag]
			if !ok {
				entry = &[2]int64{}
				perTarget[tag] = entry
			}
			switch dir {
			case "uplink", "up":
				entry[0] = stat.Value
			case "downlink", "down":
				entry[1] = stat.Value
			}
		}

		for tag, v := range perTarget {
			if err := record(now, c.category, tag, v[0], v[1]); err != nil {
				log.Printf("[Collector] record %s/%s failed: %v", c.category, tag, err)
				continue
			}
			rt.Submit(tag, v[0], v[1])
		}
		if len(perTarget) > 0 {
			log.Printf("[Collector] %s: recorded %d targets", c.category, len(perTarget))
		}
	}
}

// record 写入增量日志并累加总量
func record(ts time.Time, category, tag string, up, down int64) error {
	if up == 0 && down == 0 {
		return nil
	}
	// 增量日志
	logRow := models.TrafficLog{
		Category:      category,
		TargetName:    tag,
		UplinkDelta:   up,
		DownlinkDelta: down,
		Timestamp:     ts,
	}
	if err := db.DB.Create(&logRow).Error; err != nil {
		return err
	}

	// 累计表：存在则累加，不存在则创建（采集为单协程，无并发竞争）
	var total models.TrafficTotal
	err := db.DB.Where("category = ? AND target_name = ?", category, tag).First(&total).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		return db.DB.Create(&models.TrafficTotal{
			Category:      category,
			TargetName:    tag,
			UplinkBytes:   up,
			DownlinkBytes: down,
			TotalBytes:    up + down,
			UpdatedAt:     ts,
		}).Error
	}
	return db.DB.Model(&models.TrafficTotal{}).
		Where("id = ?", total.ID).
		Updates(map[string]interface{}{
			"uplink_bytes":   gorm.Expr("uplink_bytes + ?", up),
			"downlink_bytes": gorm.Expr("downlink_bytes + ?", down),
			"total_bytes":    gorm.Expr("total_bytes + ?", up+down),
			"updated_at":     ts,
		}).Error
}
