// Package testhelpers provides shared testing utilities for STOBot packages.
//
// This package consolidates common test helper functions to reduce duplication
// across test files and ensure consistent test infrastructure.
package testhelpers

import (
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/FracKenA/sto_news_discord_bot/internal/types"
	"github.com/bwmarrin/discordgo"

	_ "github.com/mattn/go-sqlite3"
)

// CreateTestBot creates a test instance of the Bot with an in-memory database.
// This function provides a standardized way to create test bots across all test files.
func CreateTestBot(t *testing.T) *types.Bot {
	// Create in-memory database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create basic tables for testing - match the real database schema
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS channels (
			id TEXT PRIMARY KEY,
			platforms TEXT NOT NULL DEFAULT 'pc,xbox,ps',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS posted_news (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			news_id INTEGER NOT NULL,
			channel_id TEXT NOT NULL,
			posted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(news_id, channel_id),
			FOREIGN KEY (channel_id) REFERENCES channels(id)
		);
		CREATE TABLE IF NOT EXISTS news_cache (
			id INTEGER PRIMARY KEY,
			title TEXT NOT NULL,
			summary TEXT,
			content TEXT,
			tags TEXT,
			platforms TEXT,
			updated_at DATETIME,
			thumbnail_url TEXT,
			fetched_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	config := &types.Config{
		DiscordToken: "test_token",
		PollPeriod:   30,
		PollCount:    10,
		FreshSeconds: 86400,
		MsgCount:     5,
		DatabasePath: ":memory:",
	}

	return &types.Bot{
		Session: nil, // Will be nil for testing
		DB:      db,
		Config:  config,
	}
}

// GetValidTestConfig returns a valid configuration for testing purposes.
func GetValidTestConfig() *types.Config {
	return &types.Config{
		DiscordToken: "valid_token",
		PollPeriod:   30,
		PollCount:    10,
		FreshSeconds: 86400,
		MsgCount:     5,
		DatabasePath: "/path/to/db",
	}
}

// GetInvalidTestConfigs returns a slice of invalid configurations for testing validation.
func GetInvalidTestConfigs() []struct {
	Name   string
	Config *types.Config
} {
	return []struct {
		Name   string
		Config *types.Config
	}{
		{
			Name: "empty discord token",
			Config: &types.Config{
				DiscordToken: "",
				PollPeriod:   30,
				PollCount:    10,
				FreshSeconds: 86400,
				MsgCount:     5,
				DatabasePath: "/path/to/db",
			},
		},
		{
			Name: "zero poll period",
			Config: &types.Config{
				DiscordToken: "token",
				PollPeriod:   0,
				PollCount:    10,
				FreshSeconds: 86400,
				MsgCount:     5,
				DatabasePath: "/path/to/db",
			},
		},
		{
			Name: "zero poll count",
			Config: &types.Config{
				DiscordToken: "token",
				PollPeriod:   30,
				PollCount:    0,
				FreshSeconds: 86400,
				MsgCount:     5,
				DatabasePath: "/path/to/db",
			},
		},
		{
			Name: "zero fresh seconds",
			Config: &types.Config{
				DiscordToken: "token",
				PollPeriod:   30,
				PollCount:    10,
				FreshSeconds: 0,
				MsgCount:     5,
				DatabasePath: "/path/to/db",
			},
		},
		{
			Name: "zero message count",
			Config: &types.Config{
				DiscordToken: "token",
				PollPeriod:   30,
				PollCount:    10,
				FreshSeconds: 86400,
				MsgCount:     0,
				DatabasePath: "/path/to/db",
			},
		},
		{
			Name: "empty database path",
			Config: &types.Config{
				DiscordToken: "token",
				PollPeriod:   30,
				PollCount:    10,
				FreshSeconds: 86400,
				MsgCount:     5,
				DatabasePath: "",
			},
		},
	}
}

// CreateMockDiscordSession creates a mock Discord session for testing
func CreateMockDiscordSession() *discordgo.Session {
	// Create a basic HTTP client to prevent nil pointer panics
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	session := &discordgo.Session{
		Token:  "Bot test_token",
		State:  &discordgo.State{},
		Client: client,
	}

	// Initialize rate limiter to prevent nil pointer panics
	session.Ratelimiter = discordgo.NewRatelimiter()

	return session
}
