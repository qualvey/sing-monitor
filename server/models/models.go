package models

import (
	"time"
)

// DefaultCycleDays 是用户未单独配置周期天数时的默认周期长度（天）
const DefaultCycleDays = 30

type User struct {
	ID         uint   `gorm:"primaryKey"`
	Tag        string `gorm:"uniqueIndex"`
	CreatedAt  time.Time
	CycleStart time.Time // 周期起始时间（锚点）；零值时回退到 CreatedAt
	CycleDays  int       // 周期长度（天）；<=0 时使用 DefaultCycleDays
}

// CycleDuration 返回用户的周期时长
func (u User) CycleDuration() time.Duration {
	days := u.CycleDays
	if days <= 0 {
		days = DefaultCycleDays
	}
	return time.Duration(days) * 24 * time.Hour
}

// CycleAnchor 返回周期锚点
func (u User) CycleAnchor() time.Time {
	if u.CycleStart.IsZero() {
		return u.CreatedAt
	}
	return u.CycleStart
}

// CurrentCycleWindow 返回 now 所处的当前周期窗口 [start, end)。
// 窗口自动滚动：第 n 个周期 = [anchor+n*span, anchor+(n+1)*span)，n 由当前时间推导。
func (u User) CurrentCycleWindow(now time.Time) (time.Time, time.Time) {
	anchor := u.CycleAnchor()
	span := u.CycleDuration()
	if span <= 0 {
		span = time.Duration(DefaultCycleDays) * 24 * time.Hour
	}
	n := int64(now.Sub(anchor) / span)
	if n < 0 {
		n = 0 // 未到周期起点时，从起点开始算
	}
	start := anchor.Add(time.Duration(n) * span)
	return start, start.Add(span)
}

type TrafficLog struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint `gorm:"index"`
	User      User `gorm:"foreignKey:UserID"`
	UpBytes   int64
	DownBytes int64
	Timestamp time.Time `gorm:"index"`
}

type SysStatLog struct {
	ID         uint `gorm:"primaryKey"`
	Goroutines uint32
	AllocBytes uint64
	SysBytes   uint64
	Uptime     uint32
	Timestamp  time.Time `gorm:"index"`
}
