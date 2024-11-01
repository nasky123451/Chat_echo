package middlewares

import (
	"net/http"
	"time"

	"example.com/m/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
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

func MiddlewareJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 Authorization 头部
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.RespondWithError(c, http.StatusUnauthorized, "Authorization header is required")
			c.Abort() // 终止处理
			return
		}

		// 检查 Bearer 令牌格式
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			utils.RespondWithError(c, http.StatusUnauthorized, "Invalid authorization format")
			c.Abort() // 终止处理
			return
		}

		// 提取令牌
		tokenString := authHeader[7:] // 去掉 "Bearer " 前缀

		// 解析和验证令牌
		claims, err := ParseToken(tokenString)
		if err != nil {
			utils.RespondWithError(c, http.StatusUnauthorized, "Invalid token")
			c.Abort() // 终止处理
			return
		}

		// 将解析后的用户信息添加到上下文中
		c.Set("username", claims.Username)

		// 继续处理请求
		c.Next()
	}
}
