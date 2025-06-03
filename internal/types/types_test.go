// Package types contains tests for the STOBot types package.
//
// These tests cover configuration, NewsItem helpers, and type validation.
package types

import (
	"database/sql"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		shouldError bool
	}{
		{
			name: "valid config",
			config: Config{
				DiscordToken: "valid_token",
				PollPeriod:   600,
				PollCount:    20,
				FreshSeconds: 600,
				MsgCount:     10,
				ChannelsPath: "/data/channels.txt",
				DatabasePath: "/data/stobot.db",
			},
			shouldError: false,
		},
		{
			name: "missing discord token",
			config: Config{
				PollPeriod:   600,
				PollCount:    20,
				FreshSeconds: 600,
				MsgCount:     10,
			},
			shouldError: true,
		},
		{
			name: "invalid poll period",
			config: Config{
				DiscordToken: "valid_token",
				PollPeriod:   -1,
			},
			shouldError: true,
		},
		{
			name: "invalid poll count",
			config: Config{
				DiscordToken: "valid_token",
				PollPeriod:   600,
				PollCount:    -1,
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.shouldError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestNewsItem_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		newsItem NewsItem
		expected bool
	}{
		{
			name:     "empty news item",
			newsItem: NewsItem{},
			expected: true,
		},
		{
			name: "news item with only ID",
			newsItem: NewsItem{
				ID: 12345,
			},
			expected: true,
		},
		{
			name: "news item with title",
			newsItem: NewsItem{
				ID:    12345,
				Title: "Test News",
			},
			expected: false,
		},
		{
			name: "news item with summary only",
			newsItem: NewsItem{
				Summary: "Test summary",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.newsItem.IsEmpty()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestNewsItem_HasPlatform(t *testing.T) {
	newsItem := NewsItem{
		Platforms: []string{"pc", "xbox", "ps"},
	}

	tests := []struct {
		platform string
		expected bool
	}{
		{"pc", true},
		{"xbox", true},
		{"ps", true},
		{"mobile", false},
		{"", false},
		{"PC", true}, // case insensitive
	}

	for _, tt := range tests {
		t.Run(tt.platform, func(t *testing.T) {
			result := newsItem.HasPlatform(tt.platform)
			if result != tt.expected {
				t.Errorf("HasPlatform(%q) = %v, expected %v", tt.platform, result, tt.expected)
			}
		})
	}
}

func TestNewsItem_HasTag(t *testing.T) {
	newsItem := NewsItem{
		Tags: []string{"star-trek-online", "patch-notes", "event"},
	}

	tests := []struct {
		tag      string
		expected bool
	}{
		{"star-trek-online", true},
		{"patch-notes", true},
		{"event", true},
		{"maintenance", false},
		{"", false},
		{"Star-Trek-Online", true}, // case insensitive
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			result := newsItem.HasTag(tt.tag)
			if result != tt.expected {
				t.Errorf("HasTag(%q) = %v, expected %v", tt.tag, result, tt.expected)
			}
		})
	}
}

func TestNewsItem_GetAge(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		updated  time.Time
		expected time.Duration
	}{
		{
			name:     "recent news",
			updated:  now.Add(-1 * time.Hour),
			expected: time.Hour,
		},
		{
			name:     "old news",
			updated:  now.Add(-24 * time.Hour),
			expected: 24 * time.Hour,
		},
		{
			name:     "future news",
			updated:  now.Add(1 * time.Hour),
			expected: -1 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newsItem := NewsItem{Updated: tt.updated}
			age := newsItem.GetAge()

			// Allow for small time differences due to test execution time
			diff := age - tt.expected
			if diff > time.Second || diff < -time.Second {
				t.Errorf("GetAge() = %v, expected %v (diff: %v)", age, tt.expected, diff)
			}
		})
	}
}

func TestBot_Creation(t *testing.T) {
	// Test creating a bot instance with in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	config := &Config{
		DiscordToken: "test_token",
		PollPeriod:   600,
		PollCount:    20,
		FreshSeconds: 86400,
		MsgCount:     5,
		DatabasePath: ":memory:",
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		t.Fatalf("Test configuration should be valid: %v", err)
	}

	session := &discordgo.Session{}

	bot := &Bot{
		Session: session,
		DB:      db,
		Config:  config,
	}

	// Verify bot properties
	if bot.Session != session {
		t.Error("Bot session not set correctly")
	}
	if bot.DB != db {
		t.Error("Bot database not set correctly")
	}
	if bot.Config != config {
		t.Error("Bot config not set correctly")
	}

	// Test bot with nil values (edge cases)
	nilBot := &Bot{}
	if nilBot.Session != nil {
		t.Error("Nil bot session should be nil")
	}
	if nilBot.DB != nil {
		t.Error("Nil bot database should be nil")
	}
	if nilBot.Config != nil {
		t.Error("Nil bot config should be nil")
	}
}

func TestFetchOptions_Defaults(t *testing.T) {
	options := FetchOptions{}

	// Test default values
	if options.EnablePagination != false {
		t.Error("Expected EnablePagination to default to false")
	}
	if options.ItemLimit != 0 {
		t.Error("Expected ItemLimit to default to 0")
	}
}

func TestDatabaseOptions_Defaults(t *testing.T) {
	options := DatabaseOptions{}

	// Test default values
	if options.UseBatch != false {
		t.Error("Expected UseBatch to default to false")
	}
	if options.RetryCount != 0 {
		t.Error("Expected RetryCount to default to 0")
	}
	if options.IgnoreErrors != false {
		t.Error("Expected IgnoreErrors to default to false")
	}
	if options.LogProgress != false {
		t.Error("Expected LogProgress to default to false")
	}
}

// Helper function to create a test news item
func createTestNewsItem() NewsItem {
	return NewsItem{
		ID:           12345,
		Title:        "Test News Article",
		Summary:      "This is a test news article summary",
		Tags:         []string{"star-trek-online", "test"},
		Platforms:    []string{"pc", "xbox"},
		Updated:      time.Now().Add(-1 * time.Hour),
		ThumbnailURL: "https://example.com/thumbnail.jpg",
	}
}

func TestNewsItem_String(t *testing.T) {
	newsItem := createTestNewsItem()
	str := newsItem.String()

	if str == "" {
		t.Error("String() should not return empty string")
	}

	// Should contain key information
	if !contains(str, "12345") {
		t.Error("String() should contain ID")
	}
	if !contains(str, "Test News Article") {
		t.Error("String() should contain title")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && s[:len(substr)] == substr) ||
		containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
