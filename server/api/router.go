package api

import (
	"net/http"
	"sing-monitor-server/db"
	"sing-monitor-server/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	api := r.Group("/api")
	{
		api.GET("/users", GetUsers)
		api.PUT("/users/:id/cycle", SetUserCycle)
		api.GET("/traffic/trend", GetTrafficTrend)
		api.GET("/sys/stats", GetSysStats)
	}

	return r
}

// UserOverview 概览大盘返回结构：用户信息 + 当前周期窗口 + 周期内流量
type UserOverview struct {
	ID               uint      `json:"id"`
	Tag              string    `json:"tag"`
	CreatedAt        time.Time `json:"created_at"`
	CycleDays        int       `json:"cycle_days"`
	CycleStart       time.Time `json:"cycle_start"`
	CycleEnd         time.Time `json:"cycle_end"`
	PeriodUpBytes    int64     `json:"period_up_bytes"`
	PeriodDownBytes  int64     `json:"period_down_bytes"`
	PeriodTotalBytes int64     `json:"period_total_bytes"`
}

func GetUsers(c *gin.Context) {
	var users []models.User
	if err := db.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	result := make([]UserOverview, 0, len(users))
	for _, u := range users {
		start, end := u.CurrentCycleWindow(now)

		var agg struct {
			UpSum   int64
			DownSum int64
		}
		if err := db.DB.Model(&models.TrafficLog{}).
			Select("COALESCE(SUM(up_bytes), 0) AS up_sum, COALESCE(SUM(down_bytes), 0) AS down_sum").
			Where("user_id = ? AND timestamp >= ? AND timestamp < ?", u.ID, start, end).
			Scan(&agg).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		result = append(result, UserOverview{
			ID:               u.ID,
			Tag:              u.Tag,
			CreatedAt:        u.CreatedAt,
			CycleDays:        u.CycleDays,
			CycleStart:       start,
			CycleEnd:         end,
			PeriodUpBytes:    agg.UpSum,
			PeriodDownBytes:  agg.DownSum,
			PeriodTotalBytes: agg.UpSum + agg.DownSum,
		})
	}

	c.JSON(http.StatusOK, result)
}

type SetCycleRequest struct {
	CycleDays  int        `json:"cycle_days"`
	CycleStart *time.Time `json:"cycle_start"` // 可选；不传则保持当前锚点
}

// SetUserCycle 设置用户的计费周期：起始时间 + 周期天数
func SetUserCycle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req SetCycleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.CycleDays <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cycle_days must be positive"})
		return
	}

	var user models.User
	if err := db.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	updates := map[string]interface{}{"cycle_days": req.CycleDays}
	if req.CycleStart != nil {
		updates["cycle_start"] = *req.CycleStart
	}
	if err := db.DB.Model(&user).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	db.DB.First(&user, id)
	c.JSON(http.StatusOK, user)
}

func GetTrafficTrend(c *gin.Context) {
	var logs []models.TrafficLog
	// Simple query, in a real app you'd filter by user_id and timestamp range
	userID := c.Query("user_id")
	query := db.DB.Order("timestamp desc").Limit(100)
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, logs)
}

func GetSysStats(c *gin.Context) {
	var stats []models.SysStatLog
	if err := db.DB.Order("timestamp desc").Limit(100).Find(&stats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
