package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"example.com/m/middlewares"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
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
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), // 设为过期时间
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredToken, err := token.SignedString([]byte("your-secret-key"))
	assert.NoError(t, err)

	claims, err := middlewares.ParseToken(expiredToken)

	// 确保返回错误，并且 claims 为空
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.ErrorContains(t, err, "token is expired")
}

func TestMiddlewareJWT(t *testing.T) {
	e := echo.New()
	e.Use(middlewares.MiddlewareJWT)
	e.GET("/test", func(e echo.Context) error {
		username := e.Get("username").(string)
		return e.JSON(http.StatusOK, echo.Map{"status": "success", "username": username})
	})

	t.Run("valid token", func(t *testing.T) {
		token, _ := middlewares.GenerateJWT("testuser")
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "testuser")
	})

	t.Run("missing token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("invalid token format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "InvalidTokenFormat")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("invalid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid.token.here")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
