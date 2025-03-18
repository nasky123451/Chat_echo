package middlewares

import (
	"strconv"
	"time"

	"example.com/m/metrics" // 載入你的 metrics 包
	"github.com/labstack/echo/v4"
)

func RequestMetricsMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 記錄請求開始時間
		start := time.Now()

		// 請求處理
		err := next(c)

		// 記錄請求完成後的狀態
		status := c.Response().Status
		method := c.Request().Method
		route := c.Path()
		duration := time.Since(start).Seconds()

		// 獲取 User-Agent 標頭
		userAgent := c.Request().Header.Get("User-Agent")

		// 將 status 轉換為字串
		statusStr := strconv.Itoa(status)

		// 記錄 HTTP 請求計數
		metrics.HttpRequests.WithLabelValues(method, route, statusStr, userAgent, "api").Inc()

		// 記錄 HTTP 請求延遲
		metrics.HttpDuration.WithLabelValues(method, route, "api").Observe(duration)

		// 記錄響應大小
		responseSize := float64(c.Response().Size)
		metrics.ResponseSize.WithLabelValues(method, route).Observe(responseSize)

		// 如果是成功的請求，增加活躍用戶指標
		if status == 200 {
			metrics.ActiveUsers.WithLabelValues(route).Inc()
		}

		// 處理錯誤
		if err != nil {
			c.Logger().Errorf("Request failed: %v", err)
		}

		return err
	}
}
