package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	initOnce sync.Once

	// 計數發送的聊天訊息總數
	MessageSendCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chat_message_sent_total",            // 設定指標名稱
		Help: "Total number of chat messages sent", // 提供指標的描述
	})

	// 計數接收到的聊天訊息總數
	MessageReceiveCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chat_message_received_total",            // 設定指標名稱
		Help: "Total number of chat messages received", // 提供指標的描述
	})

	// 計數用戶註冊的次數，並依照狀態來分類（例如成功或失敗）
	RegisterUserCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "register_user_counter",                   // 設定指標名稱
			Help: "Counts the number of user registrations", // 提供指標的描述
		},
		[]string{"status"}, // 使用 "status" 標籤來區分不同註冊狀態
	)

	// 計數用戶登入的次數，並依照狀態來分類（例如成功或失敗）
	LoginCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "login_counter",                    // 設定指標名稱
			Help: "Counts the number of user logins", // 提供指標的描述
		},
		[]string{"status"}, // 使用 "status" 標籤來區分不同登入狀態
	)
	// HTTP 請求計數指標
	HttpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "route", "status_code", "user_agent", "endpoint_type"},
	)

	// HTTP 請求延遲指標
	HttpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of HTTP request durations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route", "endpoint_type"},
	)

	// 用戶活躍指標
	ActiveUsers = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "active_users_total",
			Help: "Number of active users",
		},
		[]string{"route"},
	)

	// 響應大小指標
	ResponseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "Histogram of HTTP response sizes in bytes",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)
)

func InitMetrics() {
	initOnce.Do(func() {
		// 註冊 Prometheus 指標
		prometheus.MustRegister(MessageSendCounter)
		prometheus.MustRegister(MessageReceiveCounter)
		prometheus.MustRegister(RegisterUserCounter)
		prometheus.MustRegister(LoginCounter)
		prometheus.MustRegister(HttpRequests)
		prometheus.MustRegister(HttpDuration)
		prometheus.MustRegister(ActiveUsers)
		prometheus.MustRegister(ResponseSize)
	})
}
