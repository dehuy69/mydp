package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims cấu trúc cho các thông tin lưu trong JWT token
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func CreateAccessToken(username, jwtSecret string) (string, error) {
	expirationTime := time.Now().Add(time.Hour * 24 * 365 * 100) // Set the expiration time to 100 year from now

	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	encodedJWT, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}
	return encodedJWT, nil
}

func DecodeAccessToken(encodedJWT string, SECRET_KEY []byte) (map[string]interface{}, error) {
	token, err := jwt.Parse(encodedJWT, func(token *jwt.Token) (interface{}, error) {
		return SECRET_KEY, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}
