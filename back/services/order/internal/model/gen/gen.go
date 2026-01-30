package main

import (
	"log"
	"os"

	"SLGaming/back/services/order/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 数据库迁移脚本
// 使用方法：
// 1. 直接运行：go run internal/model/gen/gen.go（使用默认 DSN）
// 2. 使用环境变量：MYSQL_DSN="user:pass@tcp(host:port)/db" go run internal/model/gen/gen.go
func main() {
	// 优先从环境变量读取，否则使用默认值
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "root:root123456@tcp(120.26.29.242:3306)/SLGaming?charset=utf8mb4&parseTime=true&loc=Local"
		log.Println("使用默认 DSN，如需修改请设置环境变量 MYSQL_DSN")
	}

	log.Printf("正在连接数据库: %s", maskDSN(dsn))
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Panicf("数据库连接失败: %v", err)
	}

	log.Println("开始迁移数据库表...")
	if err := db.AutoMigrate(
		&model.Order{},
	); err != nil {
		log.Panicf("数据库迁移失败: %v", err)
		return
	}

	log.Println("数据库迁移成功！已创建/更新以下表：")
	log.Println("  - orders")
}

// maskDSN 隐藏 DSN 中的密码（用于日志输出）
func maskDSN(dsn string) string {
	// 简单处理：隐藏密码部分
	// 格式：user:password@tcp(host:port)/db
	// 这里简化处理，实际可以更精细
	if len(dsn) > 20 {
		return dsn[:10] + "***@" + dsn[len(dsn)-30:]
	}
	return "***"
}
