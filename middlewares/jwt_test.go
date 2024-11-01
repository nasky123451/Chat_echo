package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"example.com/m/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
)

func TestGenerateJWT(t *testing.T) {
	token, err := middlewares.GenerateJWT("testuser")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestParseToken_ValidToken(t *testing.T) {
	validToken, err := middlewares.GenerateJWT("testuser")
	assert.NoError(t, err)

	claims, err := middlewares.ParseToken(validToken)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "testuser", claims.Username)
}

func TestParseToken_InvalidToken(t *testing.T) {
	invalidToken := "invalid.token.here"
	claims, err := middlewares.ParseToken(invalidToken)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestParseToken_ExpiredToken(t *testing.T) {
	expiredClaims := &middlewares.Claims{
		Username: "testuser",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(-time.Hour).Unix(), // 设置令牌为过期状态
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredToken, err := token.SignedString([]byte("your-secret-key"))
	assert.NoError(t, err)

	claims, err := middlewares.ParseToken(expiredToken)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestMiddlewareJWT(t *testing.T) {
	router := gin.New()
	router.Use(middlewares.MiddlewareJWT())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success", "username": c.GetString("username")})
	})

	t.Run("valid token", func(t *testing.T) {
		token, _ := middlewares.GenerateJWT("testuser")
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "testuser")
	})

	t.Run("missing token", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid token format", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "InvalidTokenFormat")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid token", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid.token.here")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
