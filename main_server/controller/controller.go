package controller

import (
	"github.com/dehuy69/mydp/config"
	"github.com/dehuy69/mydp/main_server/service"
)

type Controller struct {
	config        *config.Config
	SQliteService *service.SQLiteService
	BadgerService *service.BadgerService
	PhoneService  *service.ParquetService
	QueueManager  *service.QueueManager
}

func NewController(config *config.Config) (*Controller, error) {
	sqliteService, err := service.NewSQLiteService(config)
	if err != nil {
		return nil, err
	}

	badgerService, err := service.NewBadgerService(config)
	if err != nil {
		return nil, err
	}

	parquetService, err := service.NewParquetService()
	if err != nil {
		return nil, err
	}

	// khởi taon internal message queue
	queueManager := service.NewQueueManager()

	return &Controller{
		config:        config,
		SQliteService: sqliteService,
		BadgerService: badgerService,
		PhoneService:  parquetService,
		QueueManager:  queueManager,
	}, nil
}

// func (ctrl *Controller) HealthCheck(c *gin.Context) {
// 	response := HealthCheckResponse{Status: "ok"}
// 	c.JSON(http.StatusOK, response)
// }
