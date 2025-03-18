package handlers

import (
	"net/http"

	"example.com/m/middlewares"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func SetupRoutes(e *echo.Echo) {

	// 記錄日誌 & 錯誤處理
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// 路由中間件，計算請求數量和延遲
	e.Use(middlewares.RequestMetricsMiddleware)

	// 路由设置
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	e.POST("/register", RegisterUser)
	e.POST("/login", LoginUser)
	e.POST("/logout", LogoutUser)

	e.GET("/ws", HandleWebSocket)

	// 使用 JWT 中间件保护以下路由
	protected := e.Group("/api")
	protected.Use(middlewares.MiddlewareJWT)
	protected.GET("/online-users", GetOnlineUsers)
	protected.GET("/chat-history", GetChatHistory)
	protected.GET("/latest-chat-date", GetLatestChatDate)

	// 添加 CORS 支持
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowHeaders:     []string{echo.HeaderAuthorization, echo.HeaderContentType},
		AllowCredentials: true,
	}))

	// 提供 React 靜態文件
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:   "./chat-app/build",
		Index:  "index.html",
		HTML5:  true,
		Browse: false,
	}))

	// React SPA 路由處理 (防止 404)
	e.GET("/*", func(c echo.Context) error {
		return c.File("./chat-app/build/index.html")
	})
}
