package models

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey"`
	Tag       string    `gorm:"uniqueIndex"`
	CreatedAt time.Time
}

type TrafficLog struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index"`
	User      User      `gorm:"foreignKey:UserID"`
	UpBytes   int64
	DownBytes int64
	Timestamp time.Time `gorm:"index"`
}

type SysStatLog struct {
	ID         uint      `gorm:"primaryKey"`
	Goroutines uint32
	AllocBytes uint64
	SysBytes   uint64
	Uptime     uint32
	Timestamp  time.Time `gorm:"index"`
}
