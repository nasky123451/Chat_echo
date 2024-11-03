package handlers

import (
	"log"
	"time"

	"example.com/m/config"
	"example.com/m/metrics"
	"example.com/m/middlewares"
	"example.com/m/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// 处理 WebSocket 连接时更新在线用户状态
func HandleWebSocket(c *gin.Context) {

	// 升级 HTTP 连接到 WebSocket
	conn, err := config.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to upgrade connection:", err)
		return
	}
	defer conn.Close()

	// 等待接收身份验证消息
	for {
		var msg map[string]string
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Error reading JSON:", err)
			break
		}

		// 处理身份验证消息
		if msg["type"] == "auth" {
			tokenString := msg["token"]
			log.Printf("Received token: %s", tokenString)

			claims, err := middlewares.ParseToken(tokenString)

			if err == nil {
				username := claims.Username
				config.Clients[conn] = username // 将用户添加到连接列表
				log.Printf("User %s connected", username)
				BroadcastUserStatus(username, true) // 广播用户上线状态

				// 更新用户在线状态到 Redis
				if err := utils.UpdateUserOnlineStatus(config.RedisClient, config.Ctx, username, true); err != nil {
					log.Println("Error updating online status in Redis:", err)
				}
			} else {
				log.Println("Could not parse claims")
				break
			}
		}

		// 处理聊天消息
		if msg["type"] == "message" {
			room := msg["room"]
			sender := msg["sender"]
			content := msg["content"]
			timeStr := msg["time"]

			msgTime, err := time.Parse(time.RFC3339, timeStr)
			if err != nil {
				log.Println("Invalid message time:", err)
				continue
			}

			filteredMessage := config.FilterMessage(content) // 使用过滤后的消息内容

			message := config.ChatMessage{
				Room:    room,
				Sender:  sender,
				Content: filteredMessage, // 使用过滤后的消息内容
				Time:    msgTime,
			}

			if err := saveMessageToDB(message); err != nil {
				log.Println("Error saving message to DB:", err)
				continue
			}

			BroadcastMessageToRoom(room, message)
		}

		// 处理登出消息
		if msg["type"] == "logout" {
			username := config.Clients[conn]
			log.Printf("User %s logging out", username)

			// 更新用户在线状态到 Redis
			if err := utils.UpdateUserOnlineStatus(config.RedisClient, config.Ctx, username, false); err != nil {
				log.Println("Error updating online status in Redis:", err)
			}

			// 广播用户下线消息
			BroadcastUserStatus(username, false)
			break // 退出循环以关闭连接
		}
	}

	// 处理用户断开连接
	username := config.Clients[conn]
	delete(config.Clients, conn)
	log.Printf("User %s disconnected", username)

	// 更新用户在线状态到 Redis
	if err := utils.UpdateUserOnlineStatus(config.RedisClient, config.Ctx, username, false); err != nil {
		log.Println("Error updating online status in Redis:", err)
	}

	// 广播用户下线消息
	BroadcastUserStatus(username, false)
}

// 广播消息到房间
func BroadcastMessageToRoom(room string, message config.ChatMessage) {
	for client, _ := range config.Clients {
		err := client.WriteJSON(gin.H{
			"type":    "message",
			"room":    message.Room,
			"sender":  message.Sender,
			"content": message.Content,
			"time":    message.Time,
		})
		if err != nil {
			config.Logger.Error("Error broadcasting message:", err)
			client.Close()
			delete(config.Clients, client)
		} else {
			metrics.MessageSendCounter.Inc() // 增加消息发送计数
		}
	}
}

// 广播用户状态
func BroadcastUserStatus(username string, online bool) {
	status := "offline"
	if online {
		status = "online"
	}

	// 使用临时切片来保存已关闭的客户端
	var closedClients []*websocket.Conn

	for client := range config.Clients {
		err := client.WriteJSON(gin.H{"type": "userStatus", "username": username, "status": status})
		if err != nil {
			log.Println("Error broadcasting user status:", err)
			// 记录关闭的客户端
			closedClients = append(closedClients, client)
		}
	}

	// 从 Clients 列表中移除已关闭的客户端
	for _, closedClient := range closedClients {
		delete(config.Clients, closedClient)
	}
}

func saveMessageToDB(message config.ChatMessage) error {
	_, err := config.PgConn.Exec(config.Ctx, "INSERT INTO chat_messages (room, sender, content, time) VALUES ($1, $2, $3, $4)",
		message.Room, message.Sender, message.Content, message.Time)
	return err
}

func saveUserDisconnectTime(username string) error {
	// 将用户断开时间记录到 PostgreSQL 中
	_, err := config.PgConn.Exec(config.Ctx, "UPDATE users SET disconnect_time = $1 WHERE username = $2", time.Now(), username)
	return err
}

// WebSocket 断开处理
func handleWebSocketDisconnect(conn *websocket.Conn, username string) {
	defer conn.Close()

	// 更新用户在线状态
	if err := utils.UpdateUserOnlineStatus(config.RedisClient, config.Ctx, username, false); err != nil {
		config.Logger.Error("Error updating online status in Redis:", err)
	}

	// 保存断开连接时间
	if err := saveUserDisconnectTime(username); err != nil {
		config.Logger.Error("Error saving disconnect time:", err)
	}

	// 广播用户状态
	BroadcastUserStatus(username, false)
}
