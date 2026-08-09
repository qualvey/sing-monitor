package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// handleWS /api/v1/ws/rt?token=xxx 实时推送
func handleWS(c *gin.Context) {
	if !app.Cfg.RealTime.Enabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "real-time disabled"})
		return
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ch := app.RT.Subscribe()
	defer app.RT.Unsubscribe(ch)

	// 读协程：检测断开
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	ticker := time.NewTicker(time.Duration(app.Cfg.RealTime.IntervalMS) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case snap := <-ch:
			if err := conn.WriteJSON(snap); err != nil {
				return
			}
		case <-ticker.C:
			// 心跳
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-done:
			return
		}
	}
}
