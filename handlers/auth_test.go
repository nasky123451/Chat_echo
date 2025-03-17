package handlers_test

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/m/config"
	"example.com/m/handlers"
	"example.com/m/metrics"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

// 初始化測試環境
func init() {
	config.Init()
	metrics.InitMetrics()
}

// Encrypt function
func encrypt(key []byte, plaintext []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	_, err = rand.Read(iv)
	if err != nil {
		return "", err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt function
func decrypt(key []byte, ciphertext string) ([]byte, error) {
	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(ciphertextBytes) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	iv := ciphertextBytes[:aes.BlockSize]
	ciphertextBytes = ciphertextBytes[aes.BlockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertextBytes, ciphertextBytes)

	return ciphertextBytes, nil
}

func TestRegisterUser(t *testing.T) {
	// Set up Gin and your routes
	e := echo.New()
	e.Logger.SetLevel(log.DEBUG)
	e.POST("/register", handlers.RegisterUser)

	// Create the request payload
	requestData := map[string]string{
		"encryptedData": "3b/f+Fd9ODHtXUIONNyRYLbZD0RuijbWBUMtYSpHEd5lFf9n/7baHS6Gfme1t/vGcd4ewBXAyFKkxi5rNK36pKiuu3FnTpp9cwAA0Zs5/099+qdIBEn6yHpdDg4NU2Du",
		"iv":            "2138ba5cb44b6906b9a5030527aab6c3",
	}

	// Marshal the request data to JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		t.Fatalf("could not marshal request data: %v", err)
	}

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a ResponseRecorder to capture the response
	w := httptest.NewRecorder()

	// Send the request to the router
	e.ServeHTTP(w, req)

	// Assert the response code and body
	assert.Equal(t, http.StatusOK, w.Code)

	// Optional: Assert the response body contains the expected message
	var responseBody map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &responseBody); err != nil {
		t.Fatalf("could not unmarshal response body: %v", err)
	}

	assert.Equal(t, "User registered", responseBody["status"])
}

func TestRegisterUserDatabaseError(t *testing.T) {
	// 每次測試前清空已註冊的指標
	prometheus.Unregister(metrics.RegisterUserCounter)
	prometheus.Unregister(metrics.LoginCounter)

	// 重新初始化 Prometheus 指標
	metrics.InitMetrics()
	// 設置 gin 引擎
	e := echo.New()
	e.Logger.SetLevel(log.DEBUG)
	e.POST("/register", handlers.RegisterUser)

	// 模擬當 PostgreSQL 連接不可用的情況
	config.PgConn = nil

	// Create invalid encrypted data
	invalidUser := map[string]string{
		"EncryptedData": "invalid-encrypted-data", // Make sure to handle this case properly in your actual test
		"IV":            "invalid-encrypted-data",
	}
	reqData, _ := json.Marshal(invalidUser)

	req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(reqData))
	if err != nil {
		t.Fatalf("Couldn't create request: %v\n", err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	// 驗證回應狀態碼
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// 驗證回應 JSON
	expectedBody := `{"error":"Decryption failed"}`
	assert.JSONEq(t, expectedBody, w.Body.String())
}

// 假設有一個函數用來刪除用戶
func deleteUser(username string) error {
	_, err := config.PgConn.Exec(config.Ctx, "DELETE FROM users WHERE username = $1", username)
	return err
}
