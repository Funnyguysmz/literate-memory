package model

import "time"

// Exam 考试表
type Exam struct {
	ID        uint      `gorm:"primaryKey"`
	ExamID    string    `gorm:"unique"` // 考试标识
	CreatedAt time.Time // 创建时间
	Status    string    // 考试状态：preparing-准备中, ongoing-进行中, finished-已结束
}

type ExamPassword struct {
	ID          uint      `gorm:"primaryKey"`
	Password    string    `gorm:"unique"`
	Distributed bool
	ExamID      string    // 考试标识
	StudentID   string    // 考生学号
	Role        string    // 用户角色：admin-管理员, student-学生
	ValidFrom   time.Time // 密码生效时间
	ValidUntil  time.Time // 密码失效时间
	Token       string    // 用于API鉴权的token
}

// AdminBackup 管理员密码备份表
type AdminBackup struct {
	ID        uint      `gorm:"primaryKey"`
	AdminID   string    `gorm:"unique"` // 管理员ID
	Password  string    // 密码
	ExamID    string    // 考试ID
	IsDefault bool      // 是否使用默认值
	CreatedAt time.Time // 创建时间
}

// DefaultAdmin 默认管理员配置
var DefaultAdmin = struct {
	AdminID  string
	Password string
	ExamID   string
}{
	AdminID:  "admin",
	Password: "admin123",
	ExamID:   "EXAM001",
}
