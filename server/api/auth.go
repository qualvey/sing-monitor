package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type loginRequest struct {
	Password string `json:"password"`
}

func handleLogin(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Password != app.Cfg.Auth.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
		return
	}
	token, err := issueToken(app.Cfg.Auth.Secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}
