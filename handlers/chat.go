package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"example.com/m/config"
	"github.com/labstack/echo/v4"
)

// 获取聊天记录
func GetChatHistory(e echo.Context) error {
	room := e.QueryParam("room")
	date := e.QueryParam("date") // 格式为 YYYY-MM-DD

	// 解析日期
	var startDate time.Time
	var endDate time.Time
	var err error

	if room == "" {
		room = "general"
	}

	if date == "" {
		// 如果没有提供日期，则使用当前日期
		startDate = time.Now().Truncate(24 * time.Hour)
		endDate = startDate.Add(24 * time.Hour)
	} else {
		startDate, err = time.Parse("2006-01-02", date)
		if err != nil {
			return e.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid date format"})
		}
		endDate = startDate.Add(24 * time.Hour)
	}

	// 查询聊天记录
	rows, err := config.PgConn.Query(config.Ctx, "SELECT sender, content, time FROM chat_messages WHERE room = $1 AND time >= $2 AND time < $3 ORDER BY time ASC", room, startDate, endDate)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, echo.Map{"error": "Error fetching chat history"})
	}
	defer rows.Close()

	var messages []config.ChatMessage
	for rows.Next() {
		var msg config.ChatMessage
		if err := rows.Scan(&msg.Sender, &msg.Content, &msg.Time); err != nil {
			return e.JSON(http.StatusInternalServerError, echo.Map{"error": "Error scanning message"})
		}
		msg.Room = room

		// msg.Content = config.FilterMessage(msg.Content) // 使用过滤后的消息内容

		messages = append(messages, msg)
	}

	// 如果没有找到消息，则返回一个状态和消息
	if len(messages) == 0 {
		return e.JSON(http.StatusOK, echo.Map{"messages": []config.ChatMessage{}, "status": "No messages found for the selected date."})
	}

	// 返回找到的消息
	return e.JSON(http.StatusOK, echo.Map{"messages": messages, "status": "Success"})
}

// 获取最新聊天日期
func GetLatestChatDate(e echo.Context) error {
	room := e.QueryParam("room") // 获取前端传来的房间参数
	var messages []config.ChatMessage
	var earliestDate *time.Time

	// 获取当前时间
	currentDate := time.Now()

	// 查询数据库中最早的聊天记录日期
	err := config.PgConn.QueryRow(config.Ctx, "SELECT MIN(time) FROM chat_messages WHERE room = $1", room).Scan(&earliestDate)
	if err != nil {
		config.Logger.Error("Error fetching earliest chat date:", err)
		return e.JSON(http.StatusInternalServerError, echo.Map{"error": "Error fetching earliest chat date"})
	}

	if earliestDate == nil {
		config.Logger.Info("No chat messages found for room:", room)
		return e.JSON(http.StatusNotFound, echo.Map{"warning": "No chat messages found"})
	}

	// 向后推一天以保证比较的完整性
	truncatedDate := earliestDate.Truncate(24 * time.Hour)
	fmt.Println(earliestDate, truncatedDate)

	// 如果没有记录，直接返回没有更多资料
	if truncatedDate.IsZero() {
		return e.JSON(http.StatusOK, echo.Map{
			"latestChatDate": "",
			"totalMessages":  "",
			"message":        "沒有更多資料",
		})
	}

	for {
		// 查询指定日期和房间的聊天记录
		rows, err := config.PgConn.Query(config.Ctx, `
			SELECT * 
			FROM chat_messages 
			WHERE DATE(time) = $1 AND room = $2 
			ORDER BY time ASC
		`, currentDate.Format("2006-01-02"), room)
		if err != nil {
			config.Logger.Error("Error fetching chat messages for date:", err)
			return e.JSON(http.StatusInternalServerError, echo.Map{"error": "Error fetching chat messages"})
		}
		defer rows.Close()

		// 将查询到的消息存入切片
		var dailyMessages []config.ChatMessage
		for rows.Next() {
			var message config.ChatMessage
			if err := rows.Scan(&message.ID, &message.Room, &message.Sender, &message.Content, &message.Time); err != nil { // 根据你的结构体字段调整
				config.Logger.Error("Error scanning message:", err)
				return e.JSON(http.StatusInternalServerError, echo.Map{"error": "Error scanning message"})
			}

			// message.Content = config.FilterMessage(message.Content) // 使用过滤后的消息内容
			dailyMessages = append(dailyMessages, message)
		}

		// 将每日的消息添加到总消息列表中
		messages = append(dailyMessages, messages...)

		// 如果当前日期的消息数量达到20，则停止
		if len(messages) >= 20 {
			break
		}

		// 向前推一天
		currentDate = currentDate.AddDate(0, 0, -1)

		// 如果已经到达最早的日期，返回没有更多资料
		if currentDate.Before(truncatedDate) {
			return e.JSON(http.StatusOK, echo.Map{
				"latestChatDate": currentDate.Format(time.RFC3339),
				"totalMessages":  messages,
				"message":        "沒有更多資料",
			})
		}
	}

	// 返回最新日期和消息总数
	return e.JSON(http.StatusOK, echo.Map{
		"latestChatDate": currentDate.Format(time.RFC3339),
		"totalMessages":  messages,
		"message":        "資料讀取完畢",
	})
}

func GetOnlineUsers(e echo.Context) error {
	// 获取在线用户列表
	onlineUsers := []string{}
	keys, err := config.RedisClient.Keys(config.Ctx, "*").Result() // 获取匹配特定模式的所有键
	if err != nil {
		return e.JSON(http.StatusInternalServerError, echo.Map{"error": "Error fetching online users"})
	}

	for _, key := range keys {
		value, err := config.RedisClient.Get(config.Ctx, key).Result() // 獲取鍵的值
		if err != nil {
			log.Printf("Error fetching value for key %s: %v", key, err)
			continue
		}
		if value == "online" { // 檢查值是否為 "online"
			onlineUsers = append(onlineUsers, key)
		}
	}

	log.Printf("Current online users: %v", onlineUsers)
	return e.JSON(http.StatusOK, echo.Map{"onlineUsers": onlineUsers})
}
