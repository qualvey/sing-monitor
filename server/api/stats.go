package api

import (
	"net/http"
	"time"

	"sing-monitor-server/db"
	"sing-monitor-server/models"

	"github.com/gin-gonic/gin"
)

// GetStats 历史统计：traffic_totals 全量（对齐原系统 StatsView）
func GetStats(c *gin.Context) {
	var totals []models.TrafficTotal
	if err := db.DB.Order("total_bytes desc").Find(&totals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, totals)
}

// PeriodSummary 概览大盘的单用户周期聚合
type PeriodSummary struct {
	TargetName      string `json:"target_name"`
	Category        string `json:"category"`
	PeriodUpBytes   int64  `json:"period_up_bytes"`
	PeriodDownBytes int64  `json:"period_down_bytes"`
	PeriodTotalBytes int64 `json:"period_total_bytes"`
	CycleStart      time.Time `json:"cycle_start"`
	CycleEnd        time.Time `json:"cycle_end"`
}

// GetStatsUsers 概览大盘：每个用户当前周期窗口内的流量（新功能核心）
func GetStatsUsers(c *gin.Context) {
	var users []models.User
	if err := db.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	now := time.Now()
	result := make([]PeriodSummary, 0, len(users))
	for _, u := range users {
		start, end := u.CurrentCycleWindow(now)
		var agg struct {
			UpSum   int64
			DownSum int64
		}
		db.DB.Model(&models.TrafficLog{}).
			Select("COALESCE(SUM(uplink_delta),0) AS up_sum, COALESCE(SUM(downlink_delta),0) AS down_sum").
			Where("category = ? AND target_name = ? AND timestamp >= ? AND timestamp < ?",
				"user", u.Email, start, end).
			Scan(&agg)
		result = append(result, PeriodSummary{
			TargetName:       u.Email,
			Category:         "user",
			PeriodUpBytes:    agg.UpSum,
			PeriodDownBytes:  agg.DownSum,
			PeriodTotalBytes: agg.UpSum + agg.DownSum,
			CycleStart:       start,
			CycleEnd:         end,
		})
	}
	c.JSON(http.StatusOK, result)
}
