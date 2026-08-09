package api

import (
	"net/http"
	"time"

	"sing-monitor-server/control"
	"sing-monitor-server/db"
	"sing-monitor-server/models"

	"github.com/gin-gonic/gin"
)

type inboundRequest struct {
	Tag                 string `json:"tag"`
	Type                string `json:"type"`
	Listen              string `json:"listen"`
	ListenPort          int64  `json:"listen_port"`
	Enable              *bool  `json:"enable"`
	ServerName          string `json:"server_name"`
	HandshakeServer     string `json:"handshake_server"`
	HandshakePort       int64  `json:"handshake_port"`
	PrivateKey          string `json:"private_key"`
	ShortID             string `json:"short_id"`
	CongestionControl   string `json:"congestion_control"`
	AuthTimeout         string `json:"auth_timeout"`
	ZeroRttHandshake    bool   `json:"zero_rtt_handshake"`
	CertificateProvider string `json:"certificate_provider"`
	ALPN                string `json:"alpn"`
}

func GetInbounds(c *gin.Context) {
	var nodes []models.InboundNode
	if err := db.DB.Order("id").Find(&nodes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, nodes)
}

func CreateInbound(c *gin.Context) {
	var req inboundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Tag == "" || req.Type == "" || req.ListenPort == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tag/type/listen_port 必填"})
		return
	}
	enable := true
	if req.Enable != nil {
		enable = *req.Enable
	}
	node := models.InboundNode{
		Tag:                 req.Tag,
		Type:                req.Type,
		Listen:              defaultStr(req.Listen, "::"),
		ListenPort:          req.ListenPort,
		Enable:              enable,
		ServerName:          req.ServerName,
		HandshakeServer:     req.HandshakeServer,
		HandshakePort:       req.HandshakePort,
		PrivateKey:          req.PrivateKey,
		ShortID:             req.ShortID,
		CongestionControl:   defaultStr(req.CongestionControl, "bbr"),
		AuthTimeout:         defaultStr(req.AuthTimeout, "3s"),
		ZeroRttHandshake:    req.ZeroRttHandshake,
		CertificateProvider: req.CertificateProvider,
		ALPN:                defaultStr(req.ALPN, "h3"),
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	if err := db.DB.Create(&node).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	regenerateAfterChange(c)
	c.JSON(http.StatusOK, node)
}

func UpdateInbound(c *gin.Context) {
	id := parseID(c.Query("id"))
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req inboundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var node models.InboundNode
	if err := db.DB.First(&node, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}

	updates := map[string]interface{}{"updated_at": time.Now()}
	if req.Tag != "" {
		updates["tag"] = req.Tag
	}
	if req.Type != "" {
		updates["type"] = req.Type
	}
	if req.Listen != "" {
		updates["listen"] = req.Listen
	}
	if req.ListenPort != 0 {
		updates["listen_port"] = req.ListenPort
	}
	if req.Enable != nil {
		updates["enable"] = *req.Enable
	}
	if req.ServerName != "" {
		updates["server_name"] = req.ServerName
	}
	if req.HandshakeServer != "" {
		updates["handshake_server"] = req.HandshakeServer
	}
	if req.HandshakePort != 0 {
		updates["handshake_port"] = req.HandshakePort
	}
	if req.PrivateKey != "" {
		updates["private_key"] = req.PrivateKey
	}
	if req.ShortID != "" {
		updates["short_id"] = req.ShortID
	}
	if req.CongestionControl != "" {
		updates["congestion_control"] = req.CongestionControl
	}
	if req.AuthTimeout != "" {
		updates["auth_timeout"] = req.AuthTimeout
	}
	if req.CertificateProvider != "" {
		updates["certificate_provider"] = req.CertificateProvider
	}
	if req.ALPN != "" {
		updates["alpn"] = req.ALPN
	}
	if err := db.DB.Model(&node).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	regenerateAfterChange(c)
	db.DB.First(&node, id)
	c.JSON(http.StatusOK, node)
}

func DeleteInbound(c *gin.Context) {
	id := parseID(c.Query("id"))
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var node models.InboundNode
	if err := db.DB.First(&node, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}
	if err := db.DB.Delete(&models.InboundNode{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	db.DB.Where("inbound_id = ?", id).Delete(&models.UserInboundBinding{})
	regenerateAfterChange(c)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

var _ = control.GenerateConfig
