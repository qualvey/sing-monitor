package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"sing-monitor-server/db"
	"sing-monitor-server/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) {
	t.Helper()
	var err error
	db.DB, err = gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.DB.AutoMigrate(&models.User{}, &models.TrafficLog{}, &models.SysStatLog{}); err != nil {
		t.Fatal(err)
	}
}

func TestGetUsersCycleTraffic(t *testing.T) {
	setupTestDB(t)
	t.Cleanup(func() {
		sqlDB, err := db.DB.DB()
		if err == nil {
			sqlDB.Close()
		}
	})

	// 用户A：30天周期，锚点 2026-08-01；当前时间模拟为 2026-08-15（窗口内）
	// 用户B：30天周期，锚点 2026-07-01；当前时间模拟为 2026-08-15（处于第二周期 [07-31, 08-30)）
	now := mustTime2("2026-08-15T12:00:00Z")

	userA := models.User{Tag: "user-a", CreatedAt: now, CycleStart: mustTime2("2026-08-01T00:00:00Z"), CycleDays: 30}
	userB := models.User{Tag: "user-b", CreatedAt: now, CycleStart: mustTime2("2026-07-01T00:00:00Z"), CycleDays: 30}
	db.DB.Create(&userA)
	db.DB.Create(&userB)

	// user-a 流量：窗口内 100MB 下载、50MB 上传；窗口外（7月）200MB 下载
	db.DB.Create(&models.TrafficLog{UserID: userA.ID, DownBytes: 100 << 20, UpBytes: 50 << 20, Timestamp: mustTime2("2026-08-10T00:00:00Z")})
	db.DB.Create(&models.TrafficLog{UserID: userA.ID, DownBytes: 200 << 20, UpBytes: 0, Timestamp: mustTime2("2026-07-20T00:00:00Z")})

	// user-b 流量：当前周期 [07-31, 08-30) 内 30MB；上一周期 500MB
	db.DB.Create(&models.TrafficLog{UserID: userB.ID, DownBytes: 30 << 20, UpBytes: 0, Timestamp: mustTime2("2026-08-05T00:00:00Z")})
	db.DB.Create(&models.TrafficLog{UserID: userB.ID, DownBytes: 500 << 20, UpBytes: 0, Timestamp: mustTime2("2026-07-05T00:00:00Z")})

	r := SetupRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	var users []UserOverview
	if err := json.Unmarshal(w.Body.Bytes(), &users); err != nil {
		t.Fatal(err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d: %s", len(users), w.Body.String())
	}

	for _, u := range users {
		switch u.Tag {
		case "user-a":
			if u.PeriodDownBytes != 100<<20 {
				t.Errorf("user-a period down = %d, want %d", u.PeriodDownBytes, 100<<20)
			}
			if u.PeriodUpBytes != 50<<20 {
				t.Errorf("user-a period up = %d, want %d", u.PeriodUpBytes, 50<<20)
			}
			if u.PeriodTotalBytes != 150<<20 {
				t.Errorf("user-a period total = %d, want %d", u.PeriodTotalBytes, 150<<20)
			}
			if !u.CycleStart.Equal(mustTime2("2026-08-01T00:00:00Z")) {
				t.Errorf("user-a cycle start = %v", u.CycleStart)
			}
		case "user-b":
			if u.PeriodDownBytes != 30<<20 {
				t.Errorf("user-b period down = %d, want %d", u.PeriodDownBytes, 30<<20)
			}
			if !u.CycleStart.Equal(mustTime2("2026-07-31T00:00:00Z")) {
				t.Errorf("user-b cycle start = %v, want 2026-07-31", u.CycleStart)
			}
			if !u.CycleEnd.Equal(mustTime2("2026-08-30T00:00:00Z")) {
				t.Errorf("user-b cycle end = %v, want 2026-08-30", u.CycleEnd)
			}
		}
	}
}

func mustTime2(s string) time.Time {
	tm, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return tm
}
