package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) DebugFunction(c *gin.Context) {
	// Get connection list of IndexService
	c.JSON(http.StatusOK, gin.H{"message": "Hello World"})
}
