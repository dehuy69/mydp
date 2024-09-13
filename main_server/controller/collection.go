package controller

import (
	"net/http"

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
