package db

import (
	"log"
	"sing-monitor-server/models"

	"gorm.io/driver/sqlite"
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
}
