package config

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	var err error
	DB, err = gorm.Open(sqlite.Open("exam_passwords.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败: ", err)
	}
}
