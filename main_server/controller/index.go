package controller

import (
	"net/http"
	"strconv"

	"github.com/dehuy69/mydp/main_server/domain"
	"github.com/dehuy69/mydp/main_server/models"
	"github.com/gin-gonic/gin"
)

// /api/workspace/<workspace-id>/collection/<collection-id>/index/create
type CreateIndexRequest struct {
	IndexName string `json:"index_name" binding:"required"`
	Fields    string `json:"fields" binding:"required"`
	IndexType string `json:"index_type" binding:"required"`
	DataType  string `json:"data_type" binding:"required"`
	IsUnique  bool   `json:"is_unique" binding:"required"`
}

func (ctrl *Controller) CreateIndexHandler(c *gin.Context) {
	collectionIDStr := c.Param("collection-id")
	collectionID, err := strconv.Atoi(collectionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collection ID"})
		return
	}

	var req CreateIndexRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	index := models.Index{
		Name:         req.IndexName,
		CollectionID: collectionID,
	}

	// Parse CreateIndexRequest vào index
	index.Fields = req.Fields
	index.IndexType = req.IndexType
	index.DataType = req.DataType
	index.IsUnique = req.IsUnique

	// Tạo index wrapper
	indexWrapper := domain.NewIndexWrapper(&index, ctrl.SQLiteCatalogService, ctrl.BadgerService, ctrl.BboltService)

	// Use indexWrapper to avoid "declared and not used" error
	if err := indexWrapper.CreateIndex(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, index)
}
