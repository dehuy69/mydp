package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) DebugFunction(c *gin.Context) {
	// Get connection list of IndexService
	connectionList := ctrl.SQLiteIndexService.GetConnectionList()
	fmt.Println(connectionList)
	c.JSON(http.StatusOK, gin.H{"message": "Hello World"})
}
