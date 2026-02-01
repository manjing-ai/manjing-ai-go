package jwtutil

import (
	"errors"
	"time"

	"manjing-ai-go/config"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT 载荷
type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

// Generate 生成 JWT
func Generate(userID int64, cfg config.JWTConfig) (string, error) {
	exp := time.Now().Add(time.Duration(cfg.ExpireDays) * 24 * time.Hour)
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(signingKey(cfg)))
}

// Parse 解析 JWT
func Parse(tokenStr string, cfg config.JWTConfig) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(signingKey(cfg)), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func signingKey(cfg config.JWTConfig) string {
	return cfg.Secret + ":" + cfg.Salt
}
