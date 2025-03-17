package handlers

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"example.com/m/middlewares"
)

func SetupRoutes(e *echo.Echo) {
	// STEP 1：讓所有 SPA 中的檔案可以在正確的路徑被找到
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:   "./chat-app/build",
		Browse: false,
	}))

	// STEP 2： serve 靜態檔案
	e.Static("/css", "public/css/")
	e.Static("/js", "public/js/")
	e.Static("/resources", "public/resources/")

	// 添加 CORS 支持
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{echo.GET, echo.POST, echo.OPTIONS},
		AllowHeaders:     []string{"Content-Type", "X-CSRF-Token", "Authorization"},
		AllowCredentials: true,
	}))

	// 处理 OPTIONS 请求
	e.OPTIONS("/register", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	// 路由设置
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	e.POST("/register", RegisterUser)
	e.POST("/login", LoginUser)
	e.POST("/logout", LogoutUser)

	e.GET("/ws", HandleWebSocket)

	// 使用 JWT 中间件保护以下路由
	protected := e.Group("/", middlewares.MiddlewareJWT())
	protected.OPTIONS("/online-users", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})
	protected.OPTIONS("/latest-chat-date", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})
	protected.GET("/online-users", GetOnlineUsers)
	protected.GET("/chat-history", GetChatHistory)
	protected.GET("/latest-chat-date", GetLatestChatDate)

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if c.Request().Method == echo.GET {
			file, _ := ioutil.ReadFile("./chat-app/build/index.html")
			etag := fmt.Sprintf("%x", md5.Sum(file))

			c.Response().Header().Set("ETag", etag)
			c.Response().Header().Set("Cache-Control", "no-cache")
			if match := c.Request().Header.Get("If-None-Match"); match != "" {
				if strings.Contains(match, etag) {
					c.NoContent(http.StatusNotModified)
					return
				}
			}
			c.Blob(http.StatusOK, "text/html; charset=utf-8", file)
			return
		}
		e.DefaultHTTPErrorHandler(err, c)
	}
}
