package control

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"sing-monitor-server/config"
	"sing-monitor-server/db"
	"sing-monitor-server/models"

	"gorm.io/gorm"
)

var jsoncComment = regexp.MustCompile(`(?m)^\s*//.*$`)

// GenerateConfig 以现有 config.json 为模板，按数据库同步用户与入站节点。
// 策略（安全优先）：
//   - 备份现有 config.json
//   - 仅更新 DB 中存在的入站节点（按 tag 匹配）的 users 数组
//   - DB 中 enable=false 的节点从 inbounds 移除
//   - config 中存在但 DB 没有的节点：保持原样，绝不误删
//   - 同步 experimental.v2ray_api.stats.users
//   - sing-box check 校验通过后写回并热重载
func GenerateConfig(cfg *config.Config) error {
	path := cfg.Control.ConfigPath
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	// 备份
	if err := backupConfig(path); err != nil {
		return err
	}

	// 解析（容忍行注释 JSONC）
	cleaned := jsoncComment.ReplaceAllString(string(raw), "")
	var root map[string]interface{}
	if err := json.Unmarshal([]byte(cleaned), &root); err != nil {
		return fmt.Errorf("parse config.json: %w", err)
	}

	// 拉取数据库状态
	dbNodes, dbUsers, err := loadDBState()
	if err != nil {
		return err
	}

	// 重写 inbounds
	inbounds, _ := root["inbounds"].([]interface{})
	newInbounds := make([]interface{}, 0, len(inbounds))
	managed := 0
	for _, item := range inbounds {
		ib, ok := item.(map[string]interface{})
		if !ok {
			newInbounds = append(newInbounds, item)
			continue
		}
		tag, _ := ib["tag"].(string)
		node, known := dbNodes[tag]
		if !known {
			// 非管理节点：原样保留
			newInbounds = append(newInbounds, item)
			continue
		}
		if !node.Enable {
			// DB 中禁用：从配置移除
			managed++
			continue
		}
		// 重写 users 数组
		users := make([]map[string]interface{}, 0, len(dbUsers[tag]))
		for _, u := range dbUsers[tag] {
			entry := map[string]interface{}{
				"name": u.Email,
				"uuid": u.UUID,
			}
			if node.Type == "vless" {
				flow := u.Flow
				if flow == "" {
					flow = "xtls-rprx-vision"
				}
				entry["flow"] = flow
			} else {
				entry["password"] = u.Password
			}
			users = append(users, entry)
		}
		ib["users"] = users
		newInbounds = append(newInbounds, ib)
		managed++
	}
	root["inbounds"] = newInbounds

	// 同步 stats users 列表
	enabledEmails := make([]string, 0, len(dbUsers))
	seen := map[string]bool{}
	for _, list := range dbUsers {
		for _, u := range list {
			if u.Enable && !seen[u.Email] {
				seen[u.Email] = true
				enabledEmails = append(enabledEmails, u.Email)
			}
		}
	}
	exp, _ := root["experimental"].(map[string]interface{})
	if exp == nil {
		exp = map[string]interface{}{}
		root["experimental"] = exp
	}
	v2api, _ := exp["v2ray_api"].(map[string]interface{})
	if v2api == nil {
		v2api = map[string]interface{}{}
		exp["v2ray_api"] = v2api
	}
	stats, _ := v2api["stats"].(map[string]interface{})
	if stats == nil {
		stats = map[string]interface{}{}
		v2api["stats"] = stats
	}
	stats["users"] = enabledEmails

	// 写出校验
	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, out, 0o644); err != nil {
		return fmt.Errorf("write tmp config: %w", err)
	}

	if err := checkConfig(cfg, tmpPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("sing-box check failed, config NOT applied: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("apply config: %w", err)
	}

	log.Printf("[Control] config regenerated: %d inbounds managed, %d stats users", managed, len(enabledEmails))

	if cfg.Control.ReloadCommand != "" {
		if err := runCommand(cfg.Control.ReloadCommand); err != nil {
			return fmt.Errorf("reload sing-box: %w", err)
		}
		log.Printf("[Control] sing-box reloaded")
	}
	return nil
}

// ImportConfig 从 config.json 导入节点/用户到数据库（同步母配置）
func ImportConfig(cfg *config.Config) error {
	raw, err := os.ReadFile(cfg.Control.ConfigPath)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	cleaned := jsoncComment.ReplaceAllString(string(raw), "")
	var root map[string]interface{}
	if err := json.Unmarshal([]byte(cleaned), &root); err != nil {
		return fmt.Errorf("parse config.json: %w", err)
	}

	inbounds, _ := root["inbounds"].([]interface{})
	imported := 0
	for _, item := range inbounds {
		ib, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		tag, _ := ib["tag"].(string)
		typ, _ := ib["type"].(string)
		if tag == "" {
			continue
		}
		// 入站节点 upsert
		node := models.InboundNode{Tag: tag}
		if err := db.DB.Where("tag = ?", tag).FirstOrCreate(&node, models.InboundNode{
			Tag:        tag,
			Type:       typ,
			Listen:     strVal(ib["listen"], "::"),
			ListenPort: int64Val(ib["listen_port"]),
			Enable:     true,
		}).Error; err != nil {
			log.Printf("[Importer] node %s: %v", tag, err)
			continue
		}
		// 节点级字段同步
		handshakePort := int64Val(ib["handshake_port"])
		if handshakePort == 0 {
			handshakePort = node.HandshakePort
		}
		listenPort := int64Val(ib["listen_port"])
		if listenPort == 0 {
			listenPort = node.ListenPort
		}
		updates := map[string]interface{}{
			"type":                 typ,
			"listen":               strVal(ib["listen"], node.Listen),
			"listen_port":          listenPort,
			"server_name":          strVal(ib["server_name"], node.ServerName),
			"handshake_server":     strVal(ib["handshake_server"], node.HandshakeServer),
			"handshake_port":       handshakePort,
			"private_key":          strVal(ib["private_key"], node.PrivateKey),
			"short_id":             strVal(ib["short_id"], node.ShortID),
			"congestion_control":   strVal(ib["congestion_control"], node.CongestionControl),
			"auth_timeout":         strVal(ib["auth_timeout"], node.AuthTimeout),
			"zero_rtt_handshake":   boolVal(ib["zero_rtt_handshake"]),
			"certificate_provider": strVal(ib["certificate_provider"], node.CertificateProvider),
			"alpn":                 strVal(ib["alpn"], node.ALPN),
		}
		if err := db.DB.Model(&models.InboundNode{}).Where("id = ?", node.ID).Updates(updates).Error; err != nil {
			log.Printf("[Importer] node %s update: %v", tag, err)
			continue
		}

		// users 数组 → users 表 upsert + 绑定
		usersArr, _ := ib["users"].([]interface{})
		for _, uItem := range usersArr {
			u, ok := uItem.(map[string]interface{})
			if !ok {
				continue
			}
			email := strVal(u["name"], "")
			if email == "" {
				continue
			}
			user := models.User{Email: email}
			if err := db.DB.Where("email = ?", email).FirstOrCreate(&user, models.User{
				Email:    email,
				UUID:     strVal(u["uuid"], ""),
				Password: strVal(u["password"], ""),
				Flow:     strVal(u["flow"], "xtls-rprx-vision"),
				Enable:   true,
			}).Error; err != nil {
				log.Printf("[Importer] user %s: %v", email, err)
				continue
			}
			// 绑定（幂等）
			var cnt int64
			db.DB.Model(&models.UserInboundBinding{}).
				Where("user_id = ? AND inbound_id = ?", user.ID, node.ID).Count(&cnt)
			if cnt == 0 {
				db.DB.Create(&models.UserInboundBinding{UserID: user.ID, InboundID: node.ID})
			}
			imported++
		}
	}
	log.Printf("[Importer] imported %d node/user bindings from %s", imported, cfg.Control.ConfigPath)
	return nil
}

// loadDBState 加载所有启用节点 + tag->绑定用户映射
func loadDBState() (map[string]models.InboundNode, map[string][]models.User, error) {
	var nodes []models.InboundNode
	if err := db.DB.Find(&nodes).Error; err != nil {
		return nil, nil, err
	}
	nodeByTag := make(map[string]models.InboundNode, len(nodes))
	for _, n := range nodes {
		nodeByTag[n.Tag] = n
	}
	usersByNode := make(map[string][]models.User)
	for _, n := range nodes {
		var users []models.User
		if err := db.DB.
			Joins("JOIN user_inbound_bindings b ON b.user_id = users.id").
			Where("b.inbound_id = ?", n.ID).
			Find(&users).Error; err != nil && err != gorm.ErrRecordNotFound {
			return nil, nil, err
		}
		usersByNode[n.Tag] = users
	}
	return nodeByTag, usersByNode, nil
}

func backupConfig(path string) error {
	dir := filepath.Dir(path)
	backupDir := filepath.Join(dir, "backup")
	_ = os.MkdirAll(backupDir, 0o755)
	name := "config.json.bak_" + time.Now().Format("20060102_150405")
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	dst := filepath.Join(backupDir, name)
	if err := os.WriteFile(dst, src, 0o644); err != nil {
		return err
	}
	log.Printf("[Control] config backed up to %s", dst)
	return nil
}

func checkConfig(cfg *config.Config, path string) error {
	if cfg.Control.CheckCommand == "" {
		return nil
	}
	cmdStr := strings.ReplaceAll(cfg.Control.CheckCommand, "%s", path)
	return runCommand(cmdStr)
}

func runCommand(cmdStr string) error {
	log.Printf("[Control] exec: %s", cmdStr)
	cmd := exec.Command("sh", "-c", cmdStr)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %v (%s)", cmdStr, err, strings.TrimSpace(string(out)))
	}
	if len(out) > 0 {
		log.Printf("[Control] %s", strings.TrimSpace(string(out)))
	}
	return nil
}

func strVal(v interface{}, def string) string {
	if s, ok := v.(string); ok && s != "" {
		return s
	}
	return def
}

func int64Val(v interface{}) int64 {
	if f, ok := v.(float64); ok {
		return int64(f)
	}
	return 0
}

func boolVal(v interface{}) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}
