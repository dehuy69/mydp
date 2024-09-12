package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func CreateAccessToken(data map[string]interface{}, expiresDelta time.Duration, SECRET_KEY []byte) (string, error) {
	claims := jwt.MapClaims{}
	for key, value := range data {
		claims[key] = value
	}

	var expire time.Time
	if expiresDelta > 0 {
		expire = time.Now().Add(expiresDelta * time.Second)
	} else {
		expire = time.Now().Add(15 * time.Minute)
	}
	claims["exp"] = expire.Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	encodedJWT, err := token.SignedString(SECRET_KEY)
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
