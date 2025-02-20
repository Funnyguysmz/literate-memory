package main

import (
	"log"
	"time"

	"GAI_test/config"
	_ "GAI_test/docs" // 这里使用下划线导入
	"GAI_test/model"
	"GAI_test/router"

	"gorm.io/gorm"
)

// @title 考试密码分发系统API
// @version 1.0
// @description 提供考试密码生成、分发和验证功能的API系统
// @host localhost:8080
// @BasePath /
func main() {
	// 初始化数据库
	config.Connect()
	// 自动迁移模型
	config.DB.AutoMigrate(&model.ExamPassword{}, &model.AdminBackup{})

	// 初始化管理员
	if err := initializeAdmin(); err != nil {
		log.Printf("初始化管理员失败: %v", err)
	} else {
		log.Println("管理员初始化检查完成")
	}

	r := router.SetupRouter()
	if err := r.Run(":8080"); err != nil {
		log.Fatal("服务启动失败：", err)
	}
}

func initializeAdmin() error {
	var count int64
	// 检查是否存在管理员备份
	config.DB.Model(&model.AdminBackup{}).Count(&count)
	if count > 0 {
		return nil
	}

	// 使用默认值创建顶级管理员
	adminPwd := model.ExamPassword{
		Password:    model.DefaultAdmin.Password,
		Distributed: false,
		ExamID:     model.DefaultAdmin.ExamID,
		StudentID:  model.DefaultAdmin.AdminID,
		Role:       "admin",
		ValidFrom:  time.Now(),
		ValidUntil: time.Now().AddDate(1, 0, 0), // 1年有效期
	}

	// 创建管理员备份
	adminBackup := model.AdminBackup{
		AdminID:   model.DefaultAdmin.AdminID,
		Password:  model.DefaultAdmin.Password,
		ExamID:    model.DefaultAdmin.ExamID,
		IsDefault: true,
		CreatedAt: time.Now(),
	}

	// 事务处理
	return config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&adminPwd).Error; err != nil {
			return err
		}
		if err := tx.Create(&adminBackup).Error; err != nil {
			return err
		}
		return nil
	})
}
