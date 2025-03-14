package model

import "time"

// User 用户表，存储基本用户信息
type User struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    string    `gorm:"uniqueIndex"` // 用户唯一标识（学号/工号）
	Name      string    // 用户姓名
	CreatedAt time.Time // 创建时间
	UpdatedAt time.Time // 更新时间
}
