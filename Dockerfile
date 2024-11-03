# 使用 Go 的官方鏡像作為建構環境
FROM golang:1.23 AS builder

# 設定工作目錄
WORKDIR /app

# 複製 go.mod 和 go.sum 文件
COPY go.mod go.sum ./

# 下載依賴包
RUN go mod download

# 複製其餘的源碼
COPY . .

# 編譯應用程式
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o myapp .

# 使用較小的鏡像來運行應用程式
FROM alpine:latest

# 安裝必要的運行時依賴（如 curl，如果需要的話）
RUN apk --no-cache add ca-certificates

# 從建構階段複製編譯好的二進位檔
COPY --from=builder /app/myapp /myapp

# 在建構階段中添加前端構建檔案的複製
COPY --from=builder /app/chat-app/build ./chat-app/build

# 在建構階段中添加測試資料的複製
COPY --from=builder /app/combined_sensitive_words.xlsx ./combined_sensitive_words.xlsx

# 暴露應用程式埠
EXPOSE 8080

# 設定容器啟動時運行的命令
CMD ["sh", "-c", "sleep 10 && /myapp"]
