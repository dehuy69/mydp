package controller

import (
	"net/http"
	"strconv"

	"github.com/dehuy69/mydp/main_server/models"
	"github.com/gin-gonic/gin"
)

type CreateCollectionRequest struct {
	Name        string `json:"name" binding:"required"`
	WorkspaceID int    `json:"workspace_id" binding:"required"`
}

func (ctrl *Controller) CreateCollectionHandler(c *gin.Context) {
	var req CreateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	collection := models.Collection{
		Name:        req.Name,
		WorkspaceID: req.WorkspaceID,
	}

	if err := ctrl.SQliteService.CreateCollection(&collection); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, collection)
}

type WriteCollectionRequest struct {
	Data map[string]interface{} `json:"data" binding:"required"`
}

func (ctrl *Controller) WriteCollectionHandler(c *gin.Context) {
	collectionIDStr := c.Param("collection-id")
	collectionID, err := strconv.Atoi(collectionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collection ID"})
		return
	}

	var req WriteCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Thêm collection ID vào dữ liệu
	req.Data["_collection_id"] = collectionID

	// Ghi dữ liệu vào queue "write-collection"
	ctrl.QueueManager.AddToQueue("write-collection", req.Data)

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
