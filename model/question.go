package model

import "time"

// QuestionType 题目类型
type QuestionType string

const (
	// QuestionTypeSingleChoice 单选题
	QuestionTypeSingleChoice QuestionType = "single_choice"
	// QuestionTypeJudgment 判断题
	QuestionTypeJudgment QuestionType = "judgment"
)

// Question 题目模型
type Question struct {
	ID          uint        `gorm:"primaryKey"`
	ExamID      string      `gorm:"index"` // 关联的考试ID
	Type        QuestionType `gorm:"type:varchar(20)"` // 题目类型：single_choice-单选题, judgment-判断题
	Content     string      // 题干内容
	OptionA     string      // 选项A（单选题必填，判断题可为空）
	OptionB     string      // 选项B（单选题必填，判断题可为空）
	OptionC     string      // 选项C（单选题必填，判断题可为空）
	OptionD     string      // 选项D（单选题必填，判断题可为空）
	Answer      string      // 正确答案：单选题为A/B/C/D，判断题为T/F
	Score       int         // 分值
	CreatedAt   time.Time   // 创建时间
	UpdatedAt   time.Time   // 更新时间
	CreatedBy   string      // 创建者ID
} 