// main.go

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dehuy69/mydp/config"
	consumer "github.com/dehuy69/mydp/main_server/consumer/write-collection"
	"github.com/dehuy69/mydp/main_server/router" // Import the new router package
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	if cfg == nil {
		log.Fatal("Failed to load configuration.")
	}

	// Initialize controller with config.Config object
	ctrl := router.InitializeController(cfg)

	// Initialize Gin router
	r := router.SetupRouter(ctrl)

	// Create server with timeout settings
	srv := &http.Server{
		Addr:    ":19450",
		Handler: r,
	}

	// Run server in a goroutine to avoid blocking signal listening
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Initialize and run write-collection consumer in a goroutine
	consumerService, err := consumer.NewWriteCollectionConsumer(ctrl.SQLiteCatalogService, ctrl.BadgerService, ctrl.QueueManager, ctrl.SQLiteIndexService)
	if err != nil {
		log.Fatalf("Failed to initialize consumer service: %v", err)
	}

	go func() {
		if err := consumerService.Start(); err != nil {
			log.Fatalf("Failed to start consumer service: %v", err)
		}
	}()

	// Listen for system signals to perform graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a context with a timeout to give the server time to finish processing ongoing requests
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
