package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"example.com/m/config"
	"example.com/m/handlers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// 初始化測試環境
func init() {
	config.Init()
}

func TestGetLatestChatDate(t *testing.T) {

	// 設置 gin 引擎
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/chat/latest-date", handlers.GetLatestChatDate)

	// 模擬有效的請求
	req, err := http.NewRequest(http.MethodGet, "/chat/latest-date?room=general", nil)
	if err != nil {
		t.Fatalf("Couldn't create request: %v\n", err)
	}

	// 使用 httptest Recorder 來模擬 HTTP 回應
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 驗證回應狀態碼
	assert.Equal(t, http.StatusOK, w.Code)

	// 驗證回應 JSON 不為空
	assert.NotEmpty(t, w.Body.String(), "Response body should not be empty")
}

func TestGetChatHistory(t *testing.T) {
	// 設置 gin 引擎
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/chat/history", handlers.GetChatHistory)

	// 模擬有效的聊天記錄請求
	validRoom := "general"
	validDate := time.Now().Format("2006-01-02")
	req, err := http.NewRequest(http.MethodGet, "/chat/history?room="+validRoom+"&date="+validDate, nil)
	if err != nil {
		t.Fatalf("Couldn't create request: %v\n", err)
	}

	// 使用 httptest Recorder 來模擬 HTTP 回應
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 驗證回應狀態碼
	assert.Equal(t, http.StatusOK, w.Code)

	// 驗證回應 JSON 不為空
	assert.NotEmpty(t, w.Body.String(), "Response body should not be empty")
}

func TestGetOnlineUsers(t *testing.T) {
	// 設置 gin 引擎
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/chat/online-users", handlers.GetOnlineUsers)

	// 模擬有效的請求
	req, err := http.NewRequest(http.MethodGet, "/chat/online-users", nil)
	if err != nil {
		t.Fatalf("Couldn't create request: %v\n", err)
	}

	// 使用 httptest Recorder 來模擬 HTTP 回應
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 驗證回應狀態碼
	assert.Equal(t, http.StatusOK, w.Code)

	// 驗證回應 JSON 不為空
	assert.NotEmpty(t, w.Body.String(), "Response body should not be empty")
}
