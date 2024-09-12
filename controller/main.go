package controller

import (
	"github.com/dehuy69/mydp/config"
	"github.com/dehuy69/mydp/service"
)

type Controller struct {
	config        *config.Config
	SQliteService *service.SQLiteService
	BadgerService *service.BadgerService
	PhoneService  *service.ParquetService
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

	return &Controller{
		config:        config,
		SQliteService: sqliteService,
		BadgerService: badgerService,
		PhoneService:  parquetService,
	}, nil
}

// func (ctrl *Controller) HealthCheck(c *gin.Context) {
// 	response := HealthCheckResponse{Status: "ok"}
// 	c.JSON(http.StatusOK, response)
// }
