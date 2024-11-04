package main

import (
	"testing"

	"example.com/m/config"
)

func TestNewAhoCorasick(t *testing.T) {
	ac := NewAhoCorasick()
	if ac == nil {
		t.Errorf("Expected a new AhoCorasick instance, got nil")
	}
}

func TestInsertAndBuild(t *testing.T) {
	ac := NewAhoCorasick()
	ac.Insert("test")
	ac.Insert("sample")
	ac.Build()

	if len(ac.root.children) == 0 {
		t.Errorf("Expected children to be populated after Insert, got %v", ac.root.children)
	}
}

func TestFilter(t *testing.T) {
	ac := NewAhoCorasick()
	ac.Insert("badword")
	ac.Build()

	message := "This is a badword in a sentence."
	results := ac.Filter(message)

	if count, exists := results["badword"]; !exists || count != 1 {
		t.Errorf("Expected 'badword' to be detected once, got %v", results)
	}
}

func TestLoadSensitiveWords(t *testing.T) {
	// Mock database connections
	config.PgConn, _ = config.InitDB()         // Replace with a mock PostgreSQL connection
	config.RedisClient, _ = config.InitRedis() // Replace with a mock Redis client

	err := loadSensitiveWords()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(sensitiveWords) == 0 {
		t.Error("Expected sensitive words to be loaded, got none")
	}
}

func TestCheckForSplitSensitiveWords(t *testing.T) {
	sensitiveWords = []string{"bad", "word"}
	message := "This is a b.a.d word."
	results := CheckForSplitSensitiveWords(message)

	if count, exists := results["bad"]; !exists || count != 1 {
		t.Errorf("Expected 'bad' to be detected once, got %v", results)
	}
	if count, exists := results["word"]; !exists || count != 1 {
		t.Errorf("Expected 'word' to be detected once, got %v", results)
	}
}

func TestAddSensitiveWord(t *testing.T) {
	word := "newbadword"
	err := addSensitiveWord(word)
	if err != nil {
		t.Errorf("Expected no error while adding sensitive word, got %v", err)
	}

	if !contains(sensitiveWords, word) {
		t.Errorf("Expected sensitive words to include '%s', got %v", word, sensitiveWords)
	}
}

func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

func TestLoadSensitiveWordsFromExcel(t *testing.T) {
	err := loadSensitiveWordsFromExcel("./combined_sensitive_words.xlsx") // Ensure this test file exists
	if err != nil {
		t.Errorf("Expected no error loading sensitive words from Excel, got %v", err)
	}

	if len(sensitiveWords) == 0 {
		t.Error("Expected sensitive words to be loaded from Excel, got none")
	}
}
