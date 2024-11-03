# Chat Application

這是一個基於 Golang 的聊天應用程式，使用 Aho-Corasick 演算法進行敏感詞過濾，並整合了 PostgreSQL 和 Redis 作為後端資料儲存解決方案。

## 資料夾結構

```plaintext
Chat/
│
├── config/                     # 配置檔案及初始化程式
│   ├── setup.go
│   ├── redis.go
│   ├── postgres.go
│   ├── logger.go
│   └── sensitive_word.go       # 敏感詞過濾處理邏輯
│
├── handlers/                   # 處理請求的邏輯，包括路由和控制器
│   ├── auth.go                 # 用戶身份驗證相關處理
│   ├── chat.go                 # 聊天功能的請求處理
│   ├── chat_test.go            # 聊天功能的單元測試
│   ├── routes.go               # 定義應用程式的路由
│   ├── websocket.go            # WebSocket 連接及相關操作處理
│   └── websocket_test.go       # WebSocket 功能的單元測試
│   
│
├── metrics/                    # 監控和度量相關功能
│   └── prometheus.go           # 整合 Prometheus 進行性能監控
│
├── middlewares/                # 中間件功能，處理請求前後的邏輯
│   ├── jwt.go                  # JWT 身份驗證的中間件實現
│   └── jwt_test.go             # JWT 中間件的單元測試
│
├── test/                       # 測試相關程式碼
│   ├── filter.go               # 敏感詞過濾邏輯
│   └── filter_test.go          # 敏感詞過濾的單元測試
│
├── utils/                      # 工具函數，包含常用的輔助函數
│   ├── error_utils.go          # 錯誤處理相關的工具函數
│   └── redis_utils.go          # Redis 相關的工具函數
│
├── main.go                     # 應用程式的入口點，啟動服務和初始化模組
├── go.mod                      # Go module 定義，管理依賴版本
└── go.sum                      # Go module 依賴清單，記錄具體的依賴版本
