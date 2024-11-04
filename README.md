# Chat Application

This is a Golang-based chat application that uses the Aho-Corasick algorithm for sensitive word filtering and integrates PostgreSQL and Redis for backend data storage. The application supports real-time messaging, user authentication, and online status management using WebSockets.

## Table of Contents
- [Features](#features)
- [Folder Structure](#folder-structure)
- [WebSocket API](#websocket-api)
  - [WebSocket Connection](#websocket-connection)
  - [Example Workflow](#example-workflow)
  - [WebSocket Message Types](#websocket-message-types)
  - [WebSocket Message Structure](#websocket-message-structure)
  - [Broadcasting User Status](#broadcasting-user-status)
  - [Error Handling](#error-handling)
- [Setup](#setup)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Running Tests](#running-tests)
- [Future Improvements](#future-improvements)

## Features

Real-time messaging: Uses WebSockets for live chat and user status updates.
Sensitive word filtering: Implements the Aho-Corasick algorithm to filter sensitive words in chat messages.
User online status: Tracks and broadcasts user online/offline statuses using Redis.
Authentication: JWT-based authentication system to secure WebSocket connections.
Historical message storage: Chat messages are saved in PostgreSQL.
Prometheus metrics: Integrated with Prometheus for monitoring application performance.

## Folder Structure

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
```

## WebSocket API

### WebSocket Connection
The WebSocket API handles user authentication, message broadcasting, and user status updates. Here’s a summary of the key features implemented in WebSocket handling:

1. Connection Upgrade: HTTP connections are upgraded to WebSocket using the Upgrader from the Gorilla WebSocket library.
2. Authentication: Once the WebSocket connection is established, users are required to send an authentication token. The token is verified using JWT middleware, and the username is extracted from the token's claims.
3. User Status Management: When a user successfully authenticates, their online status is broadcasted to all connected clients, and their status is updated in Redis.
4. Message Broadcasting: Chat messages are filtered for sensitive content, saved to PostgreSQL, and broadcasted to all users in the same chat room.
5. Logout Handling: If a user logs out or disconnects, their status is updated to offline, and this change is broadcasted to all users.
6. Redis Integration: Redis is used to keep track of online users in real-time.

### WebSocket Message Types
- Auth: For authenticating the user via a JWT token.
- Message: For sending a chat message to a room.
- Logout: For logging out and updating the user's online status.

#### WebSocket Message Structure

```json
{
  "type": "auth",
  "token": "JWT-TOKEN"
}

{
  "type": "message",
  "room": "room1",
  "sender": "user1",
  "content": "Hello, World!",
  "time": "2024-11-04T12:34:56Z"
}

{
  "type": "logout"
}
```

### Broadcasting User Status

User status updates (online/offline) are broadcasted to all connected clients when:

- A user connects or authenticates successfully.
- A user logs out or disconnects unexpectedly. Redis is used to track these statuses in real-time, ensuring that all clients receive accurate and up-to-date information about who is online.

### Error Handling
- Errors that occur during connection, authentication, message processing, or broadcasting are logged to the console.
- The WebSocket connection is properly closed when an error occurs or when the user logs out.

## Setup

### Prerequisites
- Golang 1.19+
- PostgreSQL
- Redis
- Docker and Docker Compose

## Installation

1. **Clone the repository**:

```bash
git clone https://github.com/nasky123451/Chat.git
cd chat-app
```

2. **Install Go modules**:

```bash
go mod tidy
```

3. **Configure environment variables for PostgreSQL and Redis connections**:

To configure PostgreSQL and Redis, modify the docker-compose.yml file as follows:

1. **PostgreSQL**:

- Set your PostgreSQL credentials in the `postgres` service:

```yaml
postgres:
  environment:
    POSTGRES_USER: postgres
    POSTGRES_PASSWORD: your_postgres_password
    POSTGRES_DB: your_postgres_db
```

2. **Redis**:

- Adjust the Redis environment settings in the app service to include the correct Redis host and password:

```yaml
app:
  environment:
    - REDIS_HOST=redis
    - REDIS_PASSWORD=your_redis_password
```

If you want to update additional settings, refer to the services section in your docker-compose.yml file.

4. Start the application using Docker Compose:

```bash
docker-compose up --build
```

## Running Tests
Run unit tests for sensitive word filtering, WebSocket, and JWT middleware:
```bash
go test ./handlers ./middlewares ./test
```

## Future Improvements

- Implement a rate-limiting system for chat messages to prevent spam.
- Add more complex filtering for various special characters and symbols.
- Expand Prometheus metrics to monitor WebSocket connections and message processing times.

This README outlines the main features of the chat application and how to set it up. If you have any questions or run into issues, feel free to open an issue on the GitHub repository.
