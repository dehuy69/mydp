package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dehuy69/mydp/utils"
)

// LoginPayload cấu trúc dữ liệu yêu cầu đăng nhập
type LoginPayload struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse cấu trúc dữ liệu phản hồi sau khi đăng nhập thành công
type LoginResponse struct {
	Token string `json:"token"`
}

// LoginHandler xử lý yêu cầu đăng nhập admin với Gin
func (ctrl *Controller) LoginHandler(c *gin.Context) {
	var payload LoginPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Lấy thông tin người dùng
	user, err := ctrl.SQLiteCatalogService.GetUser(payload.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Kiểm tra mật khẩu
	if !utils.CheckPasswordHash(payload.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Tạo JWT token
	token, err := utils.CreateAccessToken(user.Username, ctrl.config.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Trả về token
	c.JSON(http.StatusOK, LoginResponse{Token: token})
}
