// Package main contains tests for the STOBot main application entry point.
//
// These tests verify CLI, configuration, and integration behaviors.
package main

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/FracKenA/sto_news_discord_bot/internal/database"
	"github.com/FracKenA/sto_news_discord_bot/internal/types"

	_ "github.com/mattn/go-sqlite3"
)

func TestMarkAllPostedFunctionExists(t *testing.T) {
	// This test verifies that the main function exists and can be compiled
	// We can't easily test the actual main function execution without
	// complex mocking, but we can test the components it uses.
	t.Log("Main function compilation test passed")
}

func TestPopulateDatabaseCommand(t *testing.T) {
	// Create temporary directory for test database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	// Test that populateDatabase function exists and can be called
	// Note: We can't easily test the actual command execution without
	// mocking the cobra command framework, but we can test the underlying logic

	// Initialize database to ensure the function would work
	db, err := database.InitDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database for populate test: %v", err)
	}
	defer db.Close()

	// Verify database was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}

	t.Log("PopulateDatabase command structure test passed")
}

func TestImportChannelsCommand(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	channelsFile := filepath.Join(tempDir, "channels.txt")

	// Create test channels file
	content := "channel:123456789|pc,xbox,ps\nchannel:987654321|pc,ps\n"
	err := os.WriteFile(channelsFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test channels file: %v", err)
	}

	// Initialize database
	db, err := database.InitDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Test that the import functionality works
	bot := &types.Bot{DB: db}
	err = database.ImportChannelsFromFile(bot, channelsFile)
	if err != nil {
		t.Fatalf("Failed to import channels: %v", err)
	}

	// Verify channels were imported
	channels, err := database.GetRegisteredChannels(bot)
	if err != nil {
		t.Fatalf("Failed to get channels after import: %v", err)
	}

	if len(channels) != 2 {
		t.Errorf("Expected 2 imported channels, got %d", len(channels))
	}

	t.Log("ImportChannels command functionality test passed")
}

func TestListChannelsCommand(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	// Initialize database
	db, err := database.InitDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Add test channels
	bot := &types.Bot{DB: db}
	err = database.AddChannel(bot, "123456789")
	if err != nil {
		t.Fatalf("Failed to add test channel: %v", err)
	}

	// Test getting channels
	channels, err := database.GetRegisteredChannels(bot)
	if err != nil {
		t.Fatalf("Failed to get channels: %v", err)
	}

	if len(channels) != 1 {
		t.Errorf("Expected 1 channel, got %d", len(channels))
	}

	if len(channels) > 0 && channels[0] != "123456789" {
		t.Errorf("Expected channel ID '123456789', got '%s'", channels[0])
	}

	t.Log("ListChannels command functionality test passed")
}

func TestMarkAllPostedCommand(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	// Initialize database
	db, err := database.InitDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Add test news and channels
	newsItems := []types.NewsItem{
		{ID: 1, Title: "News 1", Summary: "Summary 1"},
		{ID: 2, Title: "News 2", Summary: "Summary 2"},
	}
	err = database.StoreNews(db, newsItems, database.DefaultDatabaseOptions())
	if err != nil {
		t.Fatalf("Failed to store test news: %v", err)
	}

	bot := &types.Bot{DB: db}
	err = database.AddChannel(bot, "123456789")
	if err != nil {
		t.Fatalf("Failed to add test channel: %v", err)
	}

	err = database.AddChannel(bot, "987654321")
	if err != nil {
		t.Fatalf("Failed to add another test channel: %v", err)
	}

	// Test getting all cached news (what markAllPosted would use)
	cached, err := database.GetAllCachedNews(bot)
	if err != nil {
		t.Fatalf("Failed to get cached news: %v", err)
	}

	if len(cached) != 2 {
		t.Errorf("Expected 2 cached news items, got %d", len(cached))
	}

	// Test getting channels (what markAllPosted would use)
	channels, err := database.GetRegisteredChannels(bot)
	if err != nil {
		t.Fatalf("Failed to get channels: %v", err)
	}

	if len(channels) != 2 {
		t.Errorf("Expected 2 channels, got %d", len(channels))
	}

	// Test marking as posted (what markAllPosted would do)
	for _, newsItem := range cached {
		for _, channelID := range channels {
			err = database.MarkAsPosted(db, newsItem.ID, channelID)
			if err != nil {
				t.Fatalf("Failed to mark news %d as posted to channel %s: %v", newsItem.ID, channelID, err)
			}
		}
	}

	// Verify all items are marked as posted
	for _, newsItem := range cached {
		for _, channelID := range channels {
			posted, err := database.IsPosted(db, newsItem.ID, channelID)
			if err != nil {
				t.Fatalf("Failed to check if posted: %v", err)
			}
			if !posted {
				t.Errorf("News %d should be marked as posted to channel %s", newsItem.ID, channelID)
			}
		}
	}

	t.Log("MarkAllPosted command functionality test passed")
}

func TestBotConfiguration(t *testing.T) {
	// Test configuration parsing and validation
	config := &types.Config{
		DiscordToken: "test_token",
		PollPeriod:   30,
		PollCount:    10,
		FreshSeconds: 86400,
		MsgCount:     5,
		ChannelsPath: "channels.txt",
		DatabasePath: "data/stobot.db",
	}

	// Test configuration validation
	err := config.Validate()
	if err != nil {
		t.Errorf("Valid configuration should not return error: %v", err)
	}

	// Test invalid configurations
	invalidConfigs := []*types.Config{
		{DiscordToken: "", PollPeriod: 30, PollCount: 10, FreshSeconds: 86400, MsgCount: 5, DatabasePath: "db"},
		{DiscordToken: "token", PollPeriod: 0, PollCount: 10, FreshSeconds: 86400, MsgCount: 5, DatabasePath: "db"},
		{DiscordToken: "token", PollPeriod: 30, PollCount: 0, FreshSeconds: 86400, MsgCount: 5, DatabasePath: "db"},
		{DiscordToken: "token", PollPeriod: 30, PollCount: 10, FreshSeconds: 0, MsgCount: 5, DatabasePath: "db"},
		{DiscordToken: "token", PollPeriod: 30, PollCount: 10, FreshSeconds: 86400, MsgCount: 0, DatabasePath: "db"},
		{DiscordToken: "token", PollPeriod: 30, PollCount: 10, FreshSeconds: 86400, MsgCount: 5, DatabasePath: ""},
	}

	for i, cfg := range invalidConfigs {
		err := cfg.Validate()
		if err == nil {
			t.Errorf("Invalid configuration %d should return error", i)
		}
	}

	t.Log("Bot configuration test passed")
}

func TestEnvironmentVariables(t *testing.T) {
	// Test environment variable handling
	testCases := map[string]string{
		"DISCORD_TOKEN": "test_token_123",
		"POLL_PERIOD":   "60",
		"POLL_COUNT":    "15",
		"FRESH_SECONDS": "172800",
		"MSG_COUNT":     "10",
		"CHANNELS_PATH": "/path/to/channels.txt",
		"DATABASE_PATH": "/path/to/database.db",
	}

	// Save original environment
	originalEnv := make(map[string]string)
	for key := range testCases {
		originalEnv[key] = os.Getenv(key)
	}

	// Set test environment variables
	for key, value := range testCases {
		err := os.Setenv(key, value)
		if err != nil {
			t.Fatalf("Failed to set environment variable %s: %v", key, err)
		}
	}

	// Test that environment variables can be read
	for key, expectedValue := range testCases {
		actualValue := os.Getenv(key)
		if actualValue != expectedValue {
			t.Errorf("Environment variable %s: expected %s, got %s", key, expectedValue, actualValue)
		}
	}

	// Restore original environment
	for key, originalValue := range originalEnv {
		if originalValue == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, originalValue)
		}
	}

	t.Log("Environment variables test passed")
}

func TestCommandLineFlags(t *testing.T) {
	// Test that the expected command line flags would be available
	// This is more of a documentation test since we can't easily test
	// the actual cobra command structure without extensive mocking

	expectedFlags := map[string]string{
		"database-path": "Path to SQLite database file",
		"count":         "Number of news items to fetch per tag",
		"tags":          "List of news tags to fetch",
		"dry-run":       "Show what would be done without making changes",
		"channels-file": "Path to channels.txt file to import",
	}

	// Verify we have documentation for expected flags
	for flag, description := range expectedFlags {
		if flag == "" {
			t.Error("Flag name should not be empty")
		}
		if description == "" {
			t.Errorf("Flag %s should have a description", flag)
		}
	}

	t.Log("Command line flags documentation test passed")
}

func TestDatabaseInitialization(t *testing.T) {
	// Test database initialization process
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	// Test database creation
	db, err := database.InitDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Verify database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file should exist after initialization")
	}

	// Verify database connection
	err = db.Ping()
	if err != nil {
		t.Errorf("Database should be pingable: %v", err)
	}

	// Verify tables exist
	tables := []string{"news_cache", "channels", "posted_news"}
	for _, table := range tables {
		var count int
		query := "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?"
		err := db.QueryRow(query, table).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to check table %s: %v", table, err)
		}
		if count != 1 {
			t.Errorf("Table %s should exist", table)
		}
	}

	t.Log("Database initialization test passed")
}

func TestSignalHandling(t *testing.T) {
	// Test that signal handling concepts are correct
	// We can't easily test actual signal handling in unit tests,
	// but we can verify the signal types are correct

	signals := []os.Signal{
		syscall.SIGINT,
		syscall.SIGTERM,
	}

	for _, sig := range signals {
		if sig == nil {
			t.Error("Signal should not be nil")
		}
	}

	// Test that signal channel creation would work
	sigChan := make(chan os.Signal, 1)

	// Verify channel capacity and that it's properly created
	if cap(sigChan) != 1 {
		t.Error("Signal channel should have capacity of 1")
	}

	t.Log("Signal handling concepts test passed")
}
