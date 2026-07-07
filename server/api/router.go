package api

import (
	"net/http"
	"sing-monitor-server/db"
	"sing-monitor-server/models"

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
		api.GET("/traffic/trend", GetTrafficTrend)
		api.GET("/sys/stats", GetSysStats)
	}

	return r
}

func GetUsers(c *gin.Context) {
	var users []models.User
	if err := db.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// For each user, we might want to return total traffic. We can do an aggregation or just return the users for now.
	// We'll keep it simple: return users
	c.JSON(http.StatusOK, users)
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
