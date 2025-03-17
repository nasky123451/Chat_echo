package middlewares

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// 生成 JWT token 的函数
func GenerateJWT(username string) (string, error) {
	claims := Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 72).Unix(), // 设置 token 过期时间为72小时
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte("your-secret-key")) // 确保将密钥替换为您的安全密钥
}

func ParseToken(tokenString string) (*Claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("your-secret-key"), nil
	})

	if err == nil && tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}

	return nil, err
}

func MiddlewareJWT() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(e echo.Context) error {
			// 获取 Authorization 头部
			authHeader := e.Request().Header.Get("Authorization")
			if authHeader == "" {
				// 使用 Echo 的返回方式來返回錯誤
				return e.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Authorization header is required",
				})
			}

			// 检查 Bearer 令牌格式
			if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
				// 使用 Echo 的返回方式來返回錯誤
				return e.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid authorization format",
				})
			}

			// 提取令牌
			tokenString := authHeader[7:] // 去掉 "Bearer " 前缀

			// 解析和验证令牌
			claims, err := ParseToken(tokenString)
			if err != nil {
				// 使用 Echo 的返回方式來返回錯誤
				return e.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid token",
				})
			}

			// 将解析后的用户信息添加到上下文中
			e.Set("username", claims.Username)

			// 继续处理请求
			return next(e)
		}
	}
}
