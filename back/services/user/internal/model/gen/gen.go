package main

import (
	"SLGaming/back/services/user/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

func main() {
	dsn := "root:root123456@tcp(120.26.29.194:3306)/SLGaming?charset=utf8mb4&parseTime=true&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Panicf("database connection failed: %v", err)
	}
	err = db.AutoMigrate(&model.User{})
	if err != nil {
		log.Panicf("database migration failed: %v", err)
		return
	}
}
