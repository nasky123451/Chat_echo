package handlers

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"example.com/m/config"
	"example.com/m/middlewares"
	"example.com/m/utils"
)

var jwtKey = []byte("secret-key")

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

// DecryptData decrypts the given encrypted data using AES.
// 解密函数
func decryptData(encryptedData string, ivHex string, secretKey string) ([]byte, error) {
	// 解码 IV
	iv, err := hex.DecodeString(ivHex)
	if err != nil {
		return nil, err
	}

	// 解码加密数据
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, err
	}

	// 创建 AES 区块
	block, err := aes.NewCipher([]byte(secretKey))
	if err != nil {
		return nil, err
	}

	// 创建 CBC 解密器
	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(ciphertext))
	mode.CryptBlocks(decrypted, ciphertext)

	// 去除填充
	decrypted = unpad(decrypted)

	return decrypted, nil
}

// 去除填充
func unpad(src []byte) []byte {
	length := len(src)
	unpadding := int(src[length-1])
	return src[:(length - unpadding)]
}

// RegisterUser handles user registration.
func RegisterUser(c *gin.Context) {
	var requestData map[string]string
	if err := c.ShouldBindJSON(&requestData); err != nil {
		config.Logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Invalid input")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Decrypt the encrypted data
	secretKey := "your-secret-key1"

	decryptedData, err := decryptData(requestData["encryptedData"], requestData["iv"], secretKey)
	if err != nil {
		config.Logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Decryption failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Decryption failed"})
		return
	}

	// Parse decrypted data into a User struct
	var user User
	err = json.Unmarshal(decryptedData, &user)
	if err != nil {
		config.Logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to parse user data")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user data"})
		return
	}

	// Check for duplicate username
	var count int
	err = config.PgConn.QueryRow(config.Ctx, "SELECT COUNT(*) FROM users WHERE username = $1", user.Username).Scan(&count)
	if err != nil {
		config.Logger.WithField("username", user.Username).Error("Error checking username")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking username"})
		return
	}
	if count > 0 {
		config.Logger.WithField("username", user.Username).Error("Username already exists")
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	// Log successful registration
	config.Logger.WithFields(logrus.Fields{
		"username": user.Username,
	}).Info("User registered successfully")

	// Hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		config.Logger.Error("Error hashing password")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}

	// Insert the new user into the database
	_, err = config.PgConn.Exec(config.Ctx, "INSERT INTO users (username, password, phone, email) VALUES ($1, $2, $3, $4)", user.Username, hash, user.Phone, user.Email)
	if err != nil {
		config.Logger.WithField("username", user.Username).Error("Error registering user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error registering user"})
		return
	}

	config.Logger.WithField("username", user.Username).Info("User registered successfully")
	config.RegisterUserCounter.WithLabelValues("success").Inc()
	c.JSON(http.StatusOK, gin.H{"status": "User registered"})
}

// 登录用户并生成 JWT
func LoginUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		config.Logger.WithField("error", err.Error()).Error("Login failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var storedHash string
	err := config.PgConn.QueryRow(config.Ctx, "SELECT password FROM users WHERE username=$1", user.Username).Scan(&storedHash)
	if err != nil {
		config.Logger.Error("Invalid username or password")
		config.LoginCounter.WithLabelValues("failure").Inc()
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	_, err = config.PgConn.Exec(config.Ctx, "UPDATE users SET time = NOW() WHERE username = $1", user.Username)
	if err != nil {
		config.Logger.WithField("error", err.Error()).Error("Failed to update time")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update login time"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(user.Password)); err != nil {
		config.Logger.Error("Invalid username or password")
		config.LoginCounter.WithLabelValues("failure").Inc()
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// 生成 JWT token
	token, err := middlewares.GenerateJWT(user.Username)
	if err != nil {
		config.Logger.Error("Error generating token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}

	config.LoginCounter.WithLabelValues("success").Inc()
	c.JSON(http.StatusOK, gin.H{"token": token}) // 返回 token 给前端
}

// 处理用户登出
func LogoutUser(c *gin.Context) {
	// 从请求的 JWT 中提取用户信息
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	claims, err := middlewares.ParseToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	username := claims.Username

	// 更新用户在线状态到 Redis
	if err := utils.UpdateUserOnlineStatus(config.RedisClient, config.Ctx, username, false); err != nil {
		log.Println("Error updating online status in Redis:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
