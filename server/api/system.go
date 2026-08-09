package api

import (
	"net/http"

	"sing-monitor-server/control"

	"github.com/gin-gonic/gin"
)

// SystemReload 从数据库重新生成 config.json 并热重载 sing-box
func SystemReload(c *gin.Context) {
	if !app.Cfg.Control.Enabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "control disabled"})
		return
	}
	if err := control.GenerateConfig(app.Cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "配置已重新生成并重载"})
}

// SystemImport 从 /etc/sing-box/config.json 同步导入节点/用户到数据库
func SystemImport(c *gin.Context) {
	if !app.Cfg.Control.Enabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "control disabled"})
		return
	}
	if err := control.ImportConfig(app.Cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "已从母配置同步导入"})
}
