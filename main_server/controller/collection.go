package controller

import (
	"net/http"
	"strconv"

	"github.com/dehuy69/mydp/main_server/domain"
	"github.com/dehuy69/mydp/main_server/models"
	"github.com/gin-gonic/gin"
)

type CreateCollectionRequest struct {
	Name string `json:"name" binding:"required"`
}

func (ctrl *Controller) CreateCollectionHandler(c *gin.Context) {
	WorkspaceIDStr := c.Param("workspace-id")
	WorkspaceID, err := strconv.Atoi(WorkspaceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	var req CreateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	collection := models.Collection{
		Name:        req.Name,
		WorkspaceID: WorkspaceID,
	}

	// collection wrapper
	collectionWrapper := domain.NewCollectionWrapper(&collection, ctrl.SQLiteCatalogService, ctrl.BadgerService, ctrl.BboltService)

	if err := collectionWrapper.CreateCollection(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, collection)
}

func (ctrl *Controller) WriteCollectionHandler(c *gin.Context) {
	collectionIDStr := c.Param("collection-id")
	collectionID, err := strconv.Atoi(collectionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collection ID"})
		return
	}

	// Retrieve collection
	collection, err := ctrl.SQLiteCatalogService.GetCollectionByID(collectionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Collection not found"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure _key in data
	if _, ok := req["_key"]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data must contain a '_key' field"})
		return
	}

	// collection wrapper
	collectionWrapper := domain.NewCollectionWrapper(collection, ctrl.SQLiteCatalogService, ctrl.BadgerService, ctrl.BboltService)

	// Kiểm tra constrain
	err = collectionWrapper.CheckIndexConstraints(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Thêm collection ID vào dữ liệu
	req["_collection_id"] = collectionID

	// Ghi dữ liệu vào queue "write-collection"
	ctrl.QueueManager.AddToQueue("write-collection", req)

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// ForceWriteCollectionHandler ghi dữ liệu vào collection mà không thông qua WAL
func (ctrl *Controller) ForceWriteCollectionHandler(c *gin.Context) {
	collectionIDStr := c.Param("collection-id")
	collectionID, err := strconv.Atoi(collectionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collection ID"})
		return
	}

	// Retrieve collection
	collection, err := ctrl.SQLiteCatalogService.GetCollectionByID(collectionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Collection not found"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure _key in data
	if _, ok := req["_key"]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data must contain a '_key' field"})
		return
	}

	// collection wrapper
	collectionWrapper := domain.NewCollectionWrapper(collection, ctrl.SQLiteCatalogService, ctrl.BadgerService, ctrl.BboltService)

	err = collectionWrapper.Write(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
