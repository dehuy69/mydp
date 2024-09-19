package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (ctrl *Controller) GetAllBadger(c *gin.Context) {
	data, err := ctrl.BadgerService.GetAllBadger()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)

}

func (ctrl *Controller) GetAllBbolt(c *gin.Context) {
	data, err := ctrl.BboltService.GetAllBbolt()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)

}

func (ctrl *Controller) GetAllQueue(c *gin.Context) {
	data := ctrl.QueueManager.GetAllCurrentQueueAndTheirFirstData()
	c.JSON(http.StatusOK, data)
}
