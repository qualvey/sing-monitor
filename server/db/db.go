package db

import (
	"log"
	"sing-monitor-server/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(dbPath string) {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
	}

	// Migrate the schema
	err = DB.AutoMigrate(&models.User{}, &models.TrafficLog{}, &models.SysStatLog{})
	if err != nil {
		log.Fatal("failed to auto migrate database")
	}

	// 存量用户回填周期字段：CycleStart 取 CreatedAt，CycleDays 用默认值
	DB.Model(&models.User{}).Where("cycle_start IS NULL").Updates(map[string]interface{}{
		"cycle_start": gorm.Expr("created_at"),
	})
	DB.Model(&models.User{}).Where("cycle_days IS NULL OR cycle_days = 0").Update("cycle_days", models.DefaultCycleDays)
}
