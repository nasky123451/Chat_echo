package main

import (
	"log"
	"strings"
	"testing"

	// 將其替換為實際的導入路徑
	"example.com/m/config"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redis/v8"
)

// 初始化 Redis 模擬
func initMockRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // 修改為你的 Redis 配置
	})
	return rdb
}

// 測試敏感詞過濾
func TestSensitiveWordFiltering(t *testing.T) {
	// 初始化資料庫連接、Redis 連接
	var err error
	// 初始化 Redis 客戶端
	rdb, err = config.InitRedis()

	// 初始化 PostgreSQL
	pgConn, err = config.InitDB()

	// 創建 PostgreSQL 模擬
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer db.Close()

	if err != nil {
		t.Fatalf("failed to connect to mock pg: %v", err)
	}
	defer pgConn.Close()

	// 設定模擬的查詢結果
	rows := sqlmock.NewRows([]string{"word"}).
		AddRow("死廢物").
		AddRow("混蛋").
		AddRow("傻逼")

	// 這裡確保 sqlmock 和 pgxpool 結合正常
	mock.ExpectQuery("SELECT word FROM sensitive_words").WillReturnRows(rows)

	// 加載敏感詞
	err = loadSensitiveWords()
	if err != nil {
		t.Fatalf("failed to load sensitive words: %v", err)
	}

	// 建立 Aho-Corasick 機器並插入敏感詞
	ac := NewAhoCorasick()
	for _, word := range sensitiveWords {
		ac.Insert(word)
	}
	ac.Build()

	// 模擬消息處理
	message := "這是一條敏感詞測試消息，包含了死廢物和混蛋。"
	results := ac.Filter(message)

	// 驗證結果
	expectedResults := map[string]int{
		"死廢物": 1,
		"混蛋":  1,
	}
	for word, expectedCount := range expectedResults {
		if count, ok := results[word]; !ok || count != expectedCount {
			t.Errorf("expected %s count: %d, got: %d", word, expectedCount, count)
		}
	}

	// 將檢測到的敏感詞替換為 *
	filteredMessage := message
	for word := range results {
		replacement := strings.Repeat("*", len(word))
		filteredMessage = strings.ReplaceAll(filteredMessage, word, replacement)
	}
	log.Println("Filtered message:", filteredMessage)
}
