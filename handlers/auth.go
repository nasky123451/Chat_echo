package handlers

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
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

// decryptData 解密前端傳輸的加密數據
func decryptData(encryptedData string, ivHex string, secretKey string) ([]byte, error) {
	iv, err := hex.DecodeString(ivHex)
	if err != nil {
		return nil, err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher([]byte(secretKey))
	if err != nil {
		return nil, err
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(ciphertext))
	mode.CryptBlocks(decrypted, ciphertext)

	decrypted = unpad(decrypted)

	return decrypted, nil
}

// unpad 移除填充字節
func unpad(src []byte) []byte {
	length := len(src)
	if length == 0 {
		return src // 防止索引錯誤
	}
	unpadding := int(src[length-1])
	if unpadding > length || unpadding == 0 {
		return src // 防止不合法的填充
	}
	return src[:(length - unpadding)]
}

// RegisterUser 註冊新用戶
func RegisterUser(e echo.Context) error {
	var requestData map[string]string
	if err := e.Bind(&requestData); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	secretKey := "your-secret-key1"
	decryptedData, err := decryptData(requestData["encryptedData"], requestData["iv"], secretKey)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Decryption failed"})
	}

	var user User
	err = json.Unmarshal(decryptedData, &user)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to parse user data"})
	}

	var count int
	err = config.PgConn.QueryRow(config.Ctx, "SELECT COUNT(*) FROM users WHERE username = $1", user.Username).Scan(&count)
	if err != nil || count > 0 {
		return e.JSON(http.StatusConflict, map[string]string{"error": "Username already exists"})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Error hashing password"})
	}

	_, err = config.PgConn.Exec(config.Ctx, "INSERT INTO users (username, password, phone, email) VALUES ($1, $2, $3, $4)", user.Username, hash, user.Phone, user.Email)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Error registering user"})
	}

	return e.JSON(http.StatusOK, map[string]string{"status": "User registered"})
}

// LoginUser 用戶登入
func LoginUser(e echo.Context) error {
	var user User
	if err := e.Bind(&user); err != nil {
		return e.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	var storedHash string
	err := config.PgConn.QueryRow(config.Ctx, "SELECT password FROM users WHERE username=$1", user.Username).Scan(&storedHash)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(user.Password)) != nil {
		return e.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid username or password"})
	}

	_, err = config.PgConn.Exec(config.Ctx, "UPDATE users SET time = NOW() WHERE username = $1", user.Username)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update login time"})
	}

	token, err := middlewares.GenerateJWT(user.Username)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Error generating token"})
	}

	return e.JSON(http.StatusOK, map[string]string{"token": token})
}

// LogoutUser 用戶登出
func LogoutUser(e echo.Context) error {
	tokenString := e.Request().Header.Get("Authorization")
	if tokenString == "" {
		return e.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	claims, err := middlewares.ParseToken(tokenString)
	if err != nil {
		return e.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
	}

	username := claims.Username
	if err := utils.UpdateUserOnlineStatus(config.RedisClient, config.Ctx, username, false); err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not update status"})
	}

	return e.JSON(http.StatusOK, map[string]string{"message": "Logged out successfully"})
}
