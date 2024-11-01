package config

import (
	"context"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"example.com/m/metrics"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var (
	RedisClient *redis.Client
	PgConn      *pgxpool.Pool
	Ctx         = context.Background()
	Upgrader    = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	Clients     = make(map[*websocket.Conn]string)
	SessionTTL  = 10 * time.Minute
	Mu          sync.Mutex
	Logger      = logrus.New()
	AuthKey     = "YOUR_GENERATED_AUTH_KEY"
	SecretKey   = "YOUR_GENERATED_SECRET_KEY"
	Log         *logrus.Logger

	// Prometheus metrics
	RegisterUserCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_registration_total",
			Help: "Total number of user registrations",
		},
		[]string{"status"},
	)

	LoginCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_login_total",
			Help: "Total number of user logins",
		},
		[]string{"status"},
	)
	MessageCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "chat_messages_total",
		Help: "Total number of chat messages sent",
	}, []string{"room"})
)

func Init() {
	var err error
	// 初始化 Redis 客戶端
	RedisClient, err = InitRedis()

	// 初始化 PostgreSQL
	PgConn, err = InitDB()

	if err := CheckAndCreateTableChat(PgConn); err != nil {
		log.Fatalf("Error checking/creating chat table: %v", err)
	}

	// 初始化 Prometheus 监控
	metrics.InitMetrics()

	// 註冊 Prometheus 指標
	prometheus.MustRegister(RegisterUserCounter)
	prometheus.MustRegister(LoginCounter)
	prometheus.MustRegister(MessageCounter)

	// 日誌配置
	Logger.SetFormatter(&logrus.JSONFormatter{})
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		Logger.Out = file
	} else {
		log.Fatal("Failed to log to file, using default stderr")
	}
}
