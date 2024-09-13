package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dehuy69/mydp/config"
	"github.com/dehuy69/mydp/controller"
	"github.com/gin-gonic/gin"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		log.Fatal("Expected 'main' or 'worker' subcommands.")
	}

	switch os.Args[1] {
	case "main":
		runMain()
	case "worker":
		runWorker()
	default:
		printHelp()
		log.Fatalf("Unknown subcommand: %s", os.Args[1])
	}
}

func runMain() {
	// Tải cấu hình
	cfg := config.LoadConfig()
	if cfg == nil {
		log.Fatal("Failed to load configuration.")
	}

	// Khởi tạo controller với đối tượng config.Config
	ctrl, err := controller.NewController(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize controller: %v", err)
	}

	// Khởi tạo router Gin
	r := gin.Default()

	publicR := r.Group("/api")
	{
		// Đăng ký route cho API đăng ký
		publicR.POST("/login", ctrl.LoginHandler)
		publicR.POST("/collection", ctrl.CreateCollectionHandler)
	}

	// Tạo server với cấu hình timeout
	srv := &http.Server{
		Addr:    ":19450",
		Handler: r,
	}

	// Chạy server trong một goroutine để không chặn việc lắng nghe tín hiệu
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Lắng nghe tín hiệu hệ thống để thực hiện graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Tạo context với timeout để server có thời gian hoàn thành các request đang xử lý
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

func runWorker() {
	// Logic cho worker
	log.Println("Worker is running...")
	// Thêm logic cho worker ở đây
}

func printHelp() {
	fmt.Println("Usage:")
	fmt.Println("  go run main.go [command]")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  main    Run the main server")
	fmt.Println("  worker  Run the worker process")
}
