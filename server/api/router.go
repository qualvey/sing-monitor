package api

import (
	"net/http"
	"strings"
	"time"

	"sing-monitor-server/config"
	"sing-monitor-server/realtime"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Ctx 供 handler 访问全局依赖
type Ctx struct {
	Cfg *config.Config
	RT  *realtime.Broadcaster
}

var app *Ctx

func SetupRouter(cfg *config.Config, rt *realtime.Broadcaster) *gin.Engine {
	app = &Ctx{Cfg: cfg, RT: rt}

	r := gin.Default()

	// CORS
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

	api := r.Group("/api/v1")
	{
		api.POST("/auth/login", handleLogin)

		authed := api.Group("")
		authed.Use(authMiddleware(cfg.Auth.Secret))
		{
			authed.GET("/stats", GetStats)
			authed.GET("/stats/users", GetStatsUsers)
			authed.GET("/users", GetUsers)
			authed.POST("/users", CreateUser)
			authed.PUT("/users/detail", UpdateUser)
			authed.DELETE("/users/detail", DeleteUser)
			authed.PUT("/users/cycle", SetUserCycle)
			authed.GET("/inbounds", GetInbounds)
			authed.POST("/inbounds", CreateInbound)
			authed.PUT("/inbounds/detail", UpdateInbound)
			authed.DELETE("/inbounds/detail", DeleteInbound)
			authed.POST("/system/reload", SystemReload)
			authed.POST("/system/import", SystemImport)
			authed.GET("/ws/rt", handleWS)
		}
	}

	// 内嵌前端（go:embed）——未命中 API 时回退
	serveEmbedded(r)

	return r
}

// ---- JWT ----

func authMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// WebSocket 支持 ?token= 查询参数
		tokenStr := ""
		if c.Query("token") != "" {
			tokenStr = c.Query("token")
		} else {
			h := c.GetHeader("Authorization")
			if strings.HasPrefix(h, "Bearer ") {
				tokenStr = strings.TrimPrefix(h, "Bearer ")
			}
		}
		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			c.Abort()
			return
		}
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		}, jwt.WithValidMethods([]string{"HS256"}))
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func issueToken(secret string) (string, error) {
	claims := jwt.MapClaims{
		"sub": "admin",
		"exp": time.Now().Add(24 * 30 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
