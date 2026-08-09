package api

import (
	"net/http"
	"strconv"
	"time"

	"sing-monitor-server/control"
	"sing-monitor-server/db"
	"sing-monitor-server/models"

	"github.com/gin-gonic/gin"
)

// UserResp 用户列表响应（对齐前端字段 + 周期统计）
type UserResp struct {
	ID              uint64     `json:"id"`
	Email           string     `json:"email"`
	UUID            string     `json:"uuid"`
	Password        string     `json:"password"`
	Flow            string     `json:"flow"`
	Enable          bool       `json:"enable"`
	TrafficLimit    int64      `json:"traffic_limit"`
	ExpireAt        *time.Time `json:"expire_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	UsedTraffic     int64      `json:"used_traffic"`
	IsOverLimit     bool       `json:"is_over_limit"`
	InboundIDs      []uint64   `json:"inbound_ids"`
	CycleStart      *time.Time `json:"cycle_start"`
	CycleDays       int        `json:"cycle_days"`
	CycleEnd        time.Time  `json:"cycle_end"`
	PeriodUpBytes   int64      `json:"period_up_bytes"`
	PeriodDownBytes int64      `json:"period_down_bytes"`
	PeriodTotalBytes int64     `json:"period_total_bytes"`
}

type userRequest struct {
	Email        string    `json:"email"`
	UUID         string    `json:"uuid"`
	Password     string    `json:"password"`
	Flow         string    `json:"flow"`
	Enable       *bool     `json:"enable"`
	TrafficLimit *int64    `json:"traffic_limit"`
	ExpireAt     *string   `json:"expire_at"`
	InboundIDs   []uint64  `json:"inbound_ids"`
	// 周期字段（新功能）
	CycleStart *string `json:"cycle_start"`
	CycleDays  *int    `json:"cycle_days"`
}

func GetUsers(c *gin.Context) {
	var users []models.User
	if err := db.DB.Order("id").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	now := time.Now()
	resp := make([]UserResp, 0, len(users))
	for _, u := range users {
		resp = append(resp, buildUserResp(u, now))
	}
	c.JSON(http.StatusOK, resp)
}

func CreateUser(c *gin.Context) {
	var req userRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email 必填"})
		return
	}

	enable := true
	if req.Enable != nil {
		enable = *req.Enable
	}
	user := models.User{
		Email:     req.Email,
		UUID:      req.UUID,
		Password:  req.Password,
		Flow:      defaultStr(req.Flow, "xtls-rprx-vision"),
		Enable:    enable,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if req.TrafficLimit != nil {
		user.TrafficLimit = *req.TrafficLimit
	}
	if req.ExpireAt != nil && *req.ExpireAt != "" {
		if t, err := parseTime(*req.ExpireAt); err == nil {
			user.ExpireAt = &t
		}
	}
	// 周期默认：创建时间 + 30 天
	cycleStart := time.Now()
	if req.CycleStart != nil && *req.CycleStart != "" {
		if t, err := parseTime(*req.CycleStart); err == nil {
			cycleStart = t
		}
	}
	user.CycleStart = &cycleStart
	user.CycleDays = models.DefaultCycleDays
	if req.CycleDays != nil && *req.CycleDays > 0 {
		user.CycleDays = *req.CycleDays
	}

	if err := db.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := replaceBindings(user.ID, req.InboundIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	regenerateAfterChange(c)
	c.JSON(http.StatusOK, buildUserResp(user, time.Now()))
}

func UpdateUser(c *gin.Context) {
	id := parseID(c.Query("id"))
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req userRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var user models.User
	if err := db.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	updates := map[string]interface{}{"updated_at": time.Now()}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.UUID != "" {
		updates["uuid"] = req.UUID
	}
	if req.Password != "" {
		updates["password"] = req.Password
	}
	if req.Flow != "" {
		updates["flow"] = req.Flow
	}
	if req.Enable != nil {
		updates["enable"] = *req.Enable
	}
	if req.TrafficLimit != nil {
		updates["traffic_limit"] = *req.TrafficLimit
	}
	if req.ExpireAt != nil {
		if *req.ExpireAt == "" {
			updates["expire_at"] = nil
		} else if t, err := parseTime(*req.ExpireAt); err == nil {
			updates["expire_at"] = t
		}
	}
	if req.CycleStart != nil {
		if *req.CycleStart == "" {
			updates["cycle_start"] = nil
		} else if t, err := parseTime(*req.CycleStart); err == nil {
			updates["cycle_start"] = t
		}
	}
	if req.CycleDays != nil && *req.CycleDays > 0 {
		updates["cycle_days"] = *req.CycleDays
	}
	if len(updates) > 1 {
		if err := db.DB.Model(&user).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	if req.InboundIDs != nil {
		if err := replaceBindings(id, req.InboundIDs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	db.DB.First(&user, id)
	regenerateAfterChange(c)
	c.JSON(http.StatusOK, buildUserResp(user, time.Now()))
}

func DeleteUser(c *gin.Context) {
	id := parseID(c.Query("id"))
	if id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var user models.User
	if err := db.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if err := db.DB.Delete(&models.User{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	db.DB.Where("user_id = ?", id).Delete(&models.UserInboundBinding{})
	regenerateAfterChange(c)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// SetUserCycle 设置用户计费周期（新功能，仅数据库，不触发配置重写）
type setCycleRequest struct {
	UserID     uint64 `json:"user_id"`
	CycleStart string `json:"cycle_start"`
	CycleDays  int    `json:"cycle_days"`
}

func SetUserCycle(c *gin.Context) {
	var req setCycleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.UserID == 0 {
		req.UserID = parseID(c.Query("user_id"))
	}
	if req.UserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id 必填"})
		return
	}
	if req.CycleDays <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cycle_days 必须大于 0"})
		return
	}
	var user models.User
	if err := db.DB.First(&user, req.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	updates := map[string]interface{}{"cycle_days": req.CycleDays}
	if req.CycleStart != "" {
		t, err := parseTime(req.CycleStart)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cycle_start 格式错误"})
			return
		}
		updates["cycle_start"] = t
	}
	if err := db.DB.Model(&user).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	db.DB.First(&user, req.UserID)
	c.JSON(http.StatusOK, buildUserResp(user, time.Now()))
}

// ---- helpers ----

func buildUserResp(u models.User, now time.Time) UserResp {
	resp := UserResp{
		ID:           u.ID,
		Email:        u.Email,
		UUID:         u.UUID,
		Password:     u.Password,
		Flow:         u.Flow,
		Enable:       u.Enable,
		TrafficLimit: u.TrafficLimit,
		ExpireAt:     u.ExpireAt,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
		CycleStart:   u.CycleStart,
		CycleDays:    u.CycleDays,
	}

	// 累计流量
	var total models.TrafficTotal
	db.DB.Where("category = ? AND target_name = ?", "user", u.Email).First(&total)
	resp.UsedTraffic = total.TotalBytes
	resp.IsOverLimit = u.IsOverLimit(total.TotalBytes)

	// 绑定节点
	var bindings []models.UserInboundBinding
	db.DB.Where("user_id = ?", u.ID).Find(&bindings)
	for _, b := range bindings {
		resp.InboundIDs = append(resp.InboundIDs, b.InboundID)
	}

	// 周期窗口内流量
	start, end := u.CurrentCycleWindow(now)
	resp.CycleEnd = end
	var agg struct {
		UpSum   int64
		DownSum int64
	}
	db.DB.Model(&models.TrafficLog{}).
		Select("COALESCE(SUM(uplink_delta),0) AS up_sum, COALESCE(SUM(downlink_delta),0) AS down_sum").
		Where("category = ? AND target_name = ? AND timestamp >= ? AND timestamp < ?",
			"user", u.Email, start, end).
		Scan(&agg)
	resp.PeriodUpBytes = agg.UpSum
	resp.PeriodDownBytes = agg.DownSum
	resp.PeriodTotalBytes = agg.UpSum + agg.DownSum
	return resp
}

func replaceBindings(userID uint64, inboundIDs []uint64) error {
	if err := db.DB.Where("user_id = ?", userID).Delete(&models.UserInboundBinding{}).Error; err != nil {
		return err
	}
	for _, ibID := range inboundIDs {
		b := models.UserInboundBinding{UserID: userID, InboundID: ibID}
		if err := db.DB.Create(&b).Error; err != nil {
			return err
		}
	}
	return nil
}

func regenerateAfterChange(c *gin.Context) {
	if app.Cfg.Control.Enabled {
		if err := control.GenerateConfig(app.Cfg); err != nil {
			c.JSON(http.StatusOK, gin.H{"ok": true, "warning": err.Error()})
			return
		}
	}
}

func parseID(s string) uint64 {
	id, _ := strconv.ParseUint(s, 10, 64)
	return id
}

func parseTime(s string) (time.Time, error) {
	layouts := []string{time.RFC3339, "2006-01-02T15:04", "2006-01-02 15:04:05", "2006-01-02"}
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, &time.ParseError{}
}

func defaultStr(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
