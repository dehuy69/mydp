package main

import (
	"log"

	"github.com/dehuy69/mydp/api"
	"github.com/dehuy69/mydp/config"
	"github.com/dehuy69/mydp/service/sqlite_service"

	"github.com/gin-gonic/gin"
)

func main() {
	// Tải cấu hình
	cfg := config.LoadConfig()
	if cfg == nil {
		log.Fatal("Failed to load configuration.")
	}

	// Khởi tạo SQLiteManager với đối tượng config.Config
	sqliteManager, err := sqlite_service.NewSQLiteManager(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize SQLite manager: %v", err)
	}
	defer sqliteManager.Close()

	// Khởi tạo router Gin
	r := gin.Default()

	// Đăng ký route cho API đăng nhập
	r.POST("/api/admin/login", func(c *gin.Context) {
		api.AdminLoginHandler(c, sqliteManager)
	})

	// Chạy server
	log.Println("Starting server on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
