package handlers

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"example.com/m/middlewares"
)

func SetupRoutes(r *gin.Engine) {

	// STEP 1：讓所有 SPA 中的檔案可以在正確的路徑被找到
	r.Use(static.Serve("/", static.LocalFile("./chat-app/build", true)))

	// STEP 2： serve 靜態檔案
	r.Static("/css", "public/css/")
	r.Static("/js", "public/js/")
	r.Static("/resources", "public/resources/")
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	// CSRF 保護
	//r.Use(gin.WrapH(csrf.Protect([]byte("32-byte-long-auth-key"), csrf.Secure(false))(r)))

	// 添加 CORS 支持
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},                                             // 替换为你的前端地址
		AllowMethods:     []string{"POST", "GET", "OPTIONS"},                        // 确保允许 OPTIONS 方法
		AllowHeaders:     []string{"Content-Type", "X-CSRF-Token", "Authorization"}, // 添加您需要的自定义头
		AllowCredentials: true,
	}))

	// 处理 OPTIONS 请求
	r.OPTIONS("/register", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*") // 允许所有源
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		c.Status(http.StatusNoContent) // 返回 204 No Content
	})

	// 路由设置
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.POST("/register", RegisterUser)
	r.POST("/login", LoginUser)
	r.POST("/logout", LogoutUser)

	r.GET("/ws", HandleWebSocket)

	// 使用 JWT 中间件保护以下路由
	protected := r.Group("/")
	protected.Use(middlewares.MiddlewareJWT())
	{
		protected.OPTIONS("/online-users", func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		})

		protected.OPTIONS("/latest-chat-date", func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		})

		protected.GET("/online-users", GetOnlineUsers)
		protected.GET("/chat-history", GetChatHistory)
		protected.GET("/latest-chat-date", GetLatestChatDate)
	}

	r.NoRoute(func(ctx *gin.Context) {
		file, _ := ioutil.ReadFile("./chat-app/build/index.html")
		etag := fmt.Sprintf("%x", md5.Sum(file)) //nolint:gosec

		ctx.Header("ETag", etag)
		ctx.Header("Cache-Control", "no-cache")
		if match := ctx.GetHeader("If-None-Match"); match != "" {
			if strings.Contains(match, etag) {
				ctx.Status(http.StatusNotModified)

				//這裡若沒 return 的話，會執行到 ctx.Data
				return
			}
		}

		ctx.Data(http.StatusOK, "text/html; charset=utf-8", file)
	})
}
