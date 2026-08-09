package db

import (
	"log"
	"time"

	"sing-monitor-server/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 连接 PostgreSQL 并做兼容迁移。
// 现有数据库表结构直接沿用；仅新增周期字段列（不破坏存量数据）。
func InitDB(dsn string) error {
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return err
	}

	// 对齐现有 schema：仅建缺失的表（traffic 等），不重建已有表
	if err := DB.AutoMigrate(
		&models.User{},
		&models.InboundNode{},
		&models.TrafficLog{},
		&models.TrafficTotal{},
		&models.UserInboundBinding{},
	); err != nil {
		return err
	}

	// 兼容迁移：为存量用户回填周期字段
	if err := backfillCycles(); err != nil {
		return err
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return nil
}

// backfillCycles 存量用户：cycle_start 回退 created_at，cycle_days 默认 30
func backfillCycles() error {
	if err := DB.Model(&models.User{}).
		Where("cycle_start IS NULL").
		Update("cycle_start", gorm.Expr("created_at")).Error; err != nil {
		return err
	}
	if err := DB.Model(&models.User{}).
		Where("cycle_days IS NULL OR cycle_days = 0").
		Update("cycle_days", models.DefaultCycleDays).Error; err != nil {
		return err
	}
	log.Printf("[DB] cycle fields backfilled")
	return nil
}
