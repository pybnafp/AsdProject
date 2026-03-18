package common

import (
	"asd/conf"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var secretKey = []byte(conf.CONFIG.ApiConfig.SecretKey)

// GenerateJWT 生成 JWT
func GenerateJWT(userID int, username string, expHours time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(expHours).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

// ParseJWT 解析 JWT
func ParseJWT(tokenString string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return &claims, nil
}
