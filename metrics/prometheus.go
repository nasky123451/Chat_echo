package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	initOnce sync.Once

	MessageSendCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chat_message_sent_total",
		Help: "Total number of chat messages sent",
	})

	MessageReceiveCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chat_message_received_total",
		Help: "Total number of chat messages received",
	})
	RegisterUserCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "register_user_counter",
			Help: "Counts the number of user registrations",
		},
		[]string{"status"},
	)
	LoginCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "login_counter",
			Help: "Counts the number of user logins",
		},
		[]string{"status"},
	)
)

func InitMetrics() {
	initOnce.Do(func() {
		// 註冊 Prometheus 指標
		prometheus.MustRegister(MessageSendCounter)
		prometheus.MustRegister(MessageReceiveCounter)
		prometheus.MustRegister(RegisterUserCounter)
		prometheus.MustRegister(LoginCounter)
		// 添加其他 Prometheus 指標的初始化代碼
	})
}
