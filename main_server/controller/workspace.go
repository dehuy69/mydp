package controller

import (
	"net/http"

	"github.com/dehuy69/mydp/main_server/domain"
	"github.com/dehuy69/mydp/main_server/models"
	"github.com/gin-gonic/gin"
)

type CreateWorkspaceRequest struct {
	Name   string `json:"name" binding:"required"`
	UserID int    `json:"user_id" binding:"required"`
}

// Create workspace
func (ctrl *Controller) CreateWorkspaceHandler(c *gin.Context) {
	var req CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get userID from JWT
	// userID := c.MustGet("userID").(int)

	workspace := models.Workspace{
		Name:    req.Name,
		OwnerID: req.UserID,
	}

	// workspace wrapper
	workspaceWrapper := domain.NewWorkspaceWrapper(&workspace, ctrl.SQLiteCatalogService, ctrl.BadgerService)

	if err := workspaceWrapper.CreateWorkspace(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, workspace)
}
