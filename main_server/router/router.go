// router/router.go

package router

import (
	"log"

	"github.com/dehuy69/mydp/config"
	"github.com/dehuy69/mydp/main_server/controller"
	"github.com/gin-gonic/gin"
)

// SetupRouter initializes the Gin router with the provided controller and returns the *gin.Engine instance.
func SetupRouter(ctrl *controller.Controller) *gin.Engine {
	// Initialize Gin router
	r := gin.Default()

	// Register routes for the API
	publicR := r.Group("/api")
	{
		publicR.POST("/login", ctrl.LoginHandler)
		publicR.POST("/collection", ctrl.CreateCollectionHandler)
		publicR.POST("/collection/:collection-id/write", ctrl.WriteCollectionHandler)
	}

	return r
}

// InitializeController initializes the controller with the configuration.
func InitializeController(cfg *config.Config) *controller.Controller {
	ctrl, err := controller.NewController(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize controller: %v", err)
	}
	return ctrl
}
