package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"

	"github.com/dehuy69/mydp/utils"
)

// Secret key dùng để tạo mã JWT
var jwtSecret = []byte("your_secret_key")

// LoginPayload cấu trúc dữ liệu yêu cầu đăng nhập
type LoginPayload struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse cấu trúc dữ liệu phản hồi sau khi đăng nhập thành công
type LoginResponse struct {
	Token string `json:"token"`
}

// Claims cấu trúc cho các thông tin lưu trong JWT token
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// LoginHandler xử lý yêu cầu đăng nhập admin với Gin
func (ctrl *Controller) LoginHandler(c *gin.Context) {
	var payload LoginPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Lấy thông tin người dùng
	user, err := ctrl.SQliteService.GetUser(payload.Username)
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
	token, err := generateJWTToken(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Trả về token
	c.JSON(http.StatusOK, LoginResponse{Token: token})
}

// generateJWTToken tạo JWT token cho người dùng
func generateJWTToken(username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
