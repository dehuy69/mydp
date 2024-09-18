package controller

import (
	"github.com/dehuy69/mydp/config"
	"github.com/dehuy69/mydp/main_server/service"
)

type Controller struct {
	config               *config.Config
	SQLiteCatalogService *service.SQLiteCatalogService
	BadgerService        *service.BadgerService
	ParquetService       *service.ParquetService
	QueueManager         *service.QueueManager
	BboltService         *service.BboltService
}

func NewController(config *config.Config) (*Controller, error) {
	sqliteCatalogService, err := service.NewSQLiteCatalogService(config)
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

	// khởi tạo internal bbolt service
	bboltService, err := service.NewBboltService(config)
	if err != nil {
		return nil, err
	}

	// khởi taon internal message queue
	queueManager := service.NewQueueManager()

	return &Controller{
		config:               config,
		SQLiteCatalogService: sqliteCatalogService,
		BadgerService:        badgerService,
		ParquetService:       parquetService,
		QueueManager:         queueManager,
		BboltService:         bboltService,
	}, nil
}

// func (ctrl *Controller) HealthCheck(c *gin.Context) {
// 	response := HealthCheckResponse{Status: "ok"}
// 	c.JSON(http.StatusOK, response)
// }
