package main

import (
	"log"

	"SLGaming/back/services/order/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 简单的本地迁移脚本：
// go run internal/model/gen/gen.go
func main() {
	dsn := "root:root123456@tcp(120.26.29.194:3306)/SLGaming?charset=utf8mb4&parseTime=true&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Panicf("database connection failed: %v", err)
	}

	if err := db.AutoMigrate(
		&model.Order{},
		&model.OrderEventOutbox{},
	); err != nil {
		log.Panicf("database migration failed: %v", err)
		return
	}
}
