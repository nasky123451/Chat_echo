package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"example.com/m/config"
	"github.com/360EntSecGroup-Skylar/excelize"
)

var sensitiveWords []string
var ac *AhoCorasick

// Aho-Corasick狀態機結構
type AhoCorasick struct {
	root     *Node
	patterns []string
}

// Node表示Aho-Corasick中的一個節點
type Node struct {
	children map[rune]*Node
	fail     *Node
	output   []string
}

// 新建Aho-Corasick
func NewAhoCorasick() *AhoCorasick {
	return &AhoCorasick{root: &Node{children: make(map[rune]*Node)}}
}

// 插入敏感詞
func (ac *AhoCorasick) Insert(pattern string) {
	node := ac.root
	for _, char := range pattern {
		if _, ok := node.children[char]; !ok {
			node.children[char] = &Node{children: make(map[rune]*Node)}
		}
		node = node.children[char]
	}
	node.output = append(node.output, pattern)
}

// 建立失敗指標
func (ac *AhoCorasick) Build() {
	queue := []*Node{ac.root}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for char, child := range current.children {
			// 設置失敗指標
			failNode := current.fail
			for failNode != nil {
				if next, ok := failNode.children[char]; ok {
					child.fail = next
					break
				}
				failNode = failNode.fail
			}
			if child.fail == nil {
				child.fail = ac.root
			}
			child.output = append(child.output, child.fail.output...)
			queue = append(queue, child)
		}
	}
}

// 用Aho-Corasick過濾消息
func (ac *AhoCorasick) Filter(content string) map[string]int {
	node := ac.root
	results := make(map[string]int)

	for _, char := range content {
		for node != ac.root && node.children[char] == nil {
			node = node.fail
		}
		node = node.children[char]

		if node == nil {
			node = ac.root
		}

		for _, pattern := range node.output {
			results[pattern]++
		}
	}

	return results
}

// 敏感詞初始化函數：從 PostgreSQL 加載敏感詞到 Redis
func loadSensitiveWords() error {
	// 清空 Redis 中舊的敏感詞
	err := config.RedisClient.Del(config.Ctx, "sensitive_words").Err()
	if err != nil {
		return err
	}

	// 從 PostgreSQL 中獲取所有敏感詞
	rows, err := config.PgConn.Query(config.Ctx, "SELECT word FROM sensitive_words")
	if err != nil {
		return err
	}
	defer rows.Close()

	// 將敏感詞加載到 Redis
	for rows.Next() {
		var word string
		if err := rows.Scan(&word); err != nil {
			return err
		}
		sensitiveWords = append(sensitiveWords, word)
		err = config.RedisClient.SAdd(config.Ctx, "sensitive_words", word).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

// CheckForSplitSensitiveWords 检查是否有拆字的敏感词
func CheckForSplitSensitiveWords(message string) map[string]int {
	results := make(map[string]int)

	for _, word := range sensitiveWords {
		// 构建拆字模式
		splitPattern := ""
		for _, char := range word {
			splitPattern += string(char) + ".*?" // 在字符之间插入正则表达式的“.*?”以匹配任意字符
		}

		// 使用正则表达式进行匹配
		re := regexp.MustCompile(splitPattern)
		if re.MatchString(message) {
			results[word]++
		}
	}
	return results
}

// 更新敏感詞列表並重新加載 Redis
func addSensitiveWord(word string) error {
	// 插入新敏感詞到 PostgreSQL
	_, err := config.PgConn.Exec(config.Ctx, "INSERT INTO sensitive_words (word) VALUES ($1) ON CONFLICT DO NOTHING", word)
	if err != nil {
		return err
	}

	// 將新詞加載到 Redis
	err = config.RedisClient.SAdd(config.Ctx, "sensitive_words", word).Err()
	if err != nil {
		return err
	}

	// 更新敏感詞列表
	sensitiveWords = append(sensitiveWords, word)
	return nil
}

// 從 Excel 文件讀取敏感詞並插入 PostgreSQL
func loadSensitiveWordsFromExcel(filePath string) error {
	// 清空舊的敏感詞
	_, err := config.PgConn.Exec(config.Ctx, "DELETE FROM sensitive_words")
	if err != nil {
		return err
	}

	// 打開 Excel 文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return err
	}

	// 讀取工作表中的所有行
	rows := f.GetRows("Sheet1") // 根據實際工作表名稱修改

	// 從第二行開始讀取（跳過標題行）
	for i, row := range rows {
		if i == 0 { // 跳過第一行
			continue
		}

		// 遍歷行中的每個詞
		for _, word := range row {
			if word != "" { // 確保詞不為空
				// 插入敏感詞到 PostgreSQL
				_, err := config.PgConn.Exec(config.Ctx, "INSERT INTO sensitive_words (word) VALUES ($1) ON CONFLICT DO NOTHING", word)
				if err != nil {
					return err
				}
				// 將新詞加載到 Redis
				err = config.RedisClient.SAdd(config.Ctx, "sensitive_words", word).Err()
				if err != nil {
					return err
				}
				// 更新敏感詞列表
				sensitiveWords = append(sensitiveWords, word)
			}
		}
	}

	return nil
}

// 在主函數中初始化資料庫連接，Redis 連接，並處理敏感詞

func InitSensitiveWordHandler() (*AhoCorasick, error) {

	// 從 Excel 文件加載敏感詞
	err := loadSensitiveWordsFromExcel("./combined_sensitive_words.xlsx") // 替換為您的文件路徑
	if err != nil {
		log.Fatalf("Error loading sensitive words from Excel: %v", err)
	}

	// 初始化時加載敏感詞
	err = loadSensitiveWords()
	if err != nil {
		fmt.Println("Error loading sensitive words:", err)
		return nil, err
	}

	// 建立 Aho-Corasick 機器並插入敏感詞
	ac := NewAhoCorasick()
	for _, word := range sensitiveWords {
		ac.Insert(word)
	}
	ac.Build()

	return ac, nil
}

// 模擬消息處理的函數，可以作為測試或使用者輸入的範例
func SimulateMessageFiltering(message string) {
	var ac *AhoCorasick

	results := ac.Filter(message)
	for word, count := range results {
		fmt.Printf("檢測到敏感詞: %s (次數: %d)\n", word, count)
	}

	// 將檢測到的敏感詞替換為 *
	filteredMessage := message
	for word := range results {
		replacement := strings.Repeat("*", len(word))
		filteredMessage = strings.ReplaceAll(filteredMessage, word, replacement)
	}
	fmt.Println("Filtered message:", filteredMessage)
}

// 过滤消息中的敏感词（包含拆字）
func FilterMessage(message string) string {
	// 使用 Aho-Corasick 检查完整的敏感词
	results := ac.Filter(message)

	// 处理拆字的敏感词
	splitResults := CheckForSplitSensitiveWords(message)
	for word, count := range splitResults {
		results[word] += count // 合并计数
		log.Printf("检测到拆字敏感词: %s (次数: %d)\n", word, count)
	}

	// 将检测到的敏感词替换为 *
	filteredMessage := message

	// 用正则表达式匹配所有非字母数字字符（包括空格、标点符号等）
	re := regexp.MustCompile(`[^a-zA-Z0-9\p{Han}]+`) // 使用 \p{Han} 匹配汉字
	// 将非字母数字字符替换为一个空格，以便正确拆分
	normalizedMessage := re.ReplaceAllString(filteredMessage, " ")
	fmt.Println(normalizedMessage)

	// 用空格拆分消息
	parts := strings.Fields(normalizedMessage)

	// 组合成一个新的字符串
	filteredMessage = strings.Join(parts, " ")

	for word, _ := range results {
		replacement := strings.Repeat("*", len(word))
		filteredMessage = strings.ReplaceAll(filteredMessage, word, replacement)
	}

	// 记录完整的敏感词检测结果
	// for word, count := range results {
	// 	if count > 0 {
	// 		log.Printf("检测到敏感词: %s (次数: %d)\n", word, count)
	// 	}
	// }

	return filteredMessage
}

func main() {
	// 初始化資料庫連接、Redis 連接
	var err error
	// 初始化 Redis 客戶端
	config.RedisClient, err = config.InitRedis()

	// 初始化 PostgreSQL
	config.PgConn, err = config.InitDB()

	if err := config.CheckAndCreateTableChat(config.PgConn); err != nil {
		log.Fatalf("Error checking/creating chat table: %v", err)
	}

	// 從 Excel 文件加載敏感詞
	err = loadSensitiveWordsFromExcel("./combined_sensitive_words.xlsx") // 替換為您的文件路徑
	if err != nil {
		log.Fatalf("Error loading sensitive words from Excel: %v", err)
	}

	// 初始化時加載敏感詞
	err = loadSensitiveWords()
	if err != nil {
		fmt.Println("Error loading sensitive words:", err)
		return
	}

	// 建立 Aho-Corasick 機器並插入敏感詞
	ac := NewAhoCorasick()
	for _, word := range sensitiveWords {
		ac.Insert(word)
	}
	ac.Build()

	// 模擬消息處理
	// 中文對話:
	dialogues := []string{
		"你聽說過那個混蛋嗎？他真是太死廢物了！",
		"是啊，我也覺得他是個廢物，總是說些屎的話。",
		"我覺得那個屎根本不值得我們浪費時間。",
		"對啊，去你媽的，別再提他了！",
		"你這混蛋，怎麼又遲到了？",
		"抱歉，我剛剛被一個不雅詞彙的討論卡住了。",
		"別再說了，我不想聽那些死胖子的話。",
		"是的，真的很無聊。",
		"你知道嗎？最近有些人真是狗屎。",
		"是啊，特別是在社交媒體上！",
	}

	// 英文對話:
	englishDialogues := []string{
		"Did you hear what that jackass said? He’s such a loser!",
		"Totally! I can’t believe he would say something so obscene.",
		"I can’t stand that bastard! Every time he speaks, I want to shut him up.",
		"Right? He’s just full of shit!",
		"You’re such a dumbass for being late!",
		"I know, I know. I was caught up in a stupid argument.",
		"Just shut up! I don’t want to hear your whore talk anymore.",
		"Fine! I’ll keep my mouth shut from now on.",
		"Seriously, those assholes on the internet need to learn some respect.",
		"I couldn’t agree more!",
	}

	allMessages := append(dialogues, englishDialogues...)

	for _, message := range allMessages {
		results := ac.Filter(message)
		for word, count := range results {
			fmt.Printf("檢測到敏感詞: %s (次數: %d)\n", word, count)
		}

		// 將檢測到的敏感詞替換為 *
		filteredMessage := message
		for word := range results {
			replacement := strings.Repeat("*", len(word))
			filteredMessage = strings.ReplaceAll(filteredMessage, word, replacement)
		}
		fmt.Println("Filtered message:", filteredMessage)
	}
}
