package middlewares

import (
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// JWT Secret Key
var jwtSecret = []byte("your-secret-key")

// Claims 结构体
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// 生成 JWT Token
func GenerateJWT(username string) (string, error) {
	claims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)), // 72小时过期
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// 解析 Token
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, err
}

// JWT 认证中间件
func MiddlewareJWT(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")

		if authHeader == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, map[string]string{"error": "Authorization header is required"})
		}

		// 检查 Bearer 令牌格式
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return echo.NewHTTPError(http.StatusUnauthorized, map[string]string{"error": "Invalid authorization format"})
		}

		// 解析 Token
		claims, err := ParseToken(parts[1])
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
		}

		// 将用户信息存入上下文
		c.Set("username", claims.Username)

		return next(c)
	}
}
