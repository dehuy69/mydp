package main

import (
	"log"
	"os"

	"github.com/dehuy69/mydp/config"
	"github.com/dehuy69/mydp/controller"

	"github.com/gin-gonic/gin"
)

func main() {
	// Tải cấu hình
	cfg := config.LoadConfig()
	if cfg == nil {
		log.Fatal("Failed to load configuration.")
	}

	//Khởi tạo controller với đối tượng config.Config
	controller, err := controller.NewController(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize controller: %v", err)
	}

	// Khởi tạo router Gin
	r := gin.Default()

	publicR := r.Group("/api")
	{
		// Đăng ký route cho API đăng ký
		publicR.POST("/login", controller.LoginHandler)
		publicR.POST("/collection", controller.CreateCollectionHandler)
	}

	// Chạy server
	log.Println("Starting server on :19450...")
	if err := r.Run(":19450"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// createDataFolder tạo folder ./data nếu chưa tồn tại
func createDataFolder() error {
	const dataFolder = "./data"
	if _, err := os.Stat(dataFolder); os.IsNotExist(err) {
		if err := os.Mkdir(dataFolder, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}
