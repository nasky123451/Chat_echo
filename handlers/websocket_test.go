package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"example.com/m/config"
	"example.com/m/handlers"
	"example.com/m/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)

	// 清空 Clients 以确保每次测试的独立性
	config.Clients = make(map[*websocket.Conn]string)
}

// 测试 HandleWebSocket 函数
func TestHandleWebSocket(t *testing.T) {
	router := gin.Default()
	router.GET("/ws", handlers.HandleWebSocket) // 使用 handlers 包中的 HandleWebSocket 函数

	// 创建 WebSocket 请求
	req, err := http.NewRequest(http.MethodGet, "/ws", nil)
	if err != nil {
		t.Fatalf("Couldn't create request: %v\n", err)
	}

	// 设置必要的标头
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")

	// 使用 gin 的 ResponseWriter
	w := httptest.NewRecorder()

	// 使用 gin.Context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	conn, err := config.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		t.Fatalf("Couldn't upgrade to WebSocket: %v\n", err)
	}
	defer conn.Close()

	// 生成有效的 JWT token
	token, _ := middlewares.GenerateJWT("testuser")

	// 测试身份验证消息
	authMsg := map[string]string{
		"type":  "auth",
		"token": token, // 使用有效的测试 token
	}

	if err := conn.WriteJSON(authMsg); err != nil {
		t.Fatalf("Couldn't send auth message: %v\n", err)
	}

	// 读取响应，验证连接是否成功
	var msg map[string]interface{}
	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatalf("Couldn't read connection success message: %v\n", err)
	}

	assert.Equal(t, "userStatus", msg["type"])
	assert.Equal(t, "testuser", msg["username"])
	assert.Equal(t, "online", msg["status"])

	// 模拟一条聊天消息
	chatMsg := map[string]string{
		"type":    "message",
		"room":    "general",
		"sender":  "testuser",
		"content": "Hello, World!",
		"time":    time.Now().Format(time.RFC3339),
	}

	if err := conn.WriteJSON(chatMsg); err != nil {
		t.Fatalf("Couldn't send chat message: %v\n", err)
	}
	// 读取响应，这里假设消息会被广播给所有连接的用户
	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatalf("Couldn't read broadcast message: %v\n", err)
	}

	assert.Equal(t, "message", msg["type"])
	assert.Equal(t, "general", msg["room"])
	assert.Equal(t, "testuser", msg["sender"])
	assert.Equal(t, "Hello, World!", msg["content"])

	// 测试登出消息
	logoutMsg := map[string]string{
		"type": "logout",
	}

	if err := conn.WriteJSON(logoutMsg); err != nil {
		t.Fatalf("Couldn't send logout message: %v\n", err)
	}

	// 读取响应，确认用户状态已更新
	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatalf("Couldn't read logout broadcast message: %v\n", err)
	}

	assert.Equal(t, "userStatus", msg["type"])
	assert.Equal(t, "testuser", msg["username"])
	assert.Equal(t, "offline", msg["status"])
}

// 测试处理无效 token 的情况
func TestHandleWebSocketInvalidToken(t *testing.T) {

	router := gin.Default()
	router.GET("/ws", handlers.HandleWebSocket) // 使用 handlers 包中的 HandleWebSocket 函数

	req, err := http.NewRequest(http.MethodGet, "/ws", nil)
	if err != nil {
		t.Fatalf("Couldn't create request: %v\n", err)
	}

	// 设置必要的标头
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")

	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	conn, err := config.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		t.Fatalf("Couldn't upgrade to WebSocket: %v\n", err)
	}
	defer conn.Close()

	// 发送无效的身份验证消息
	authMsg := map[string]string{
		"type":  "auth",
		"token": "invalid-token",
	}

	if err := conn.WriteJSON(authMsg); err != nil {
		t.Fatalf("Couldn't send auth message: %v\n", err)
	}

	// 读取响应，应该处理登出并关闭连接
	if _, _, err := conn.ReadMessage(); err == nil {
		t.Fatal("Expected an error due to invalid token, but got none.")
	}
}
