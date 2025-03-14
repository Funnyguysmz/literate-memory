package model

import (
	"time"
)

// ScreenStreamSession 屏幕流会话信息
type ScreenStreamSession struct {
	ID           uint      `gorm:"primaryKey"`
	SessionID    string    `gorm:"uniqueIndex"` // 会话唯一标识
	UserID       string    // 用户ID
	ExamID       string    // 考试ID
	StartTime    time.Time // 开始时间
	EndTime      time.Time // 结束时间
	IsActive     bool      // 是否活跃
	StreamType   string    // 流类型：interval-定时截屏, realtime-实时直播
	IntervalSecs int       // 截屏间隔秒数（仅用于interval类型）
}

// ScreenCapture 屏幕截图记录
type ScreenCapture struct {
	ID         uint      `gorm:"primaryKey"`
	SessionID  string    // 关联的会话ID
	CaptureTime time.Time // 截图时间
	ImagePath   string    // 图片存储路径（如果需要保存）
} 