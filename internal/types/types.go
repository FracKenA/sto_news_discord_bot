// Package types defines core types and configuration structures for STOBot.
//
// It includes Bot, Config, NewsItem, and related helpers.
package types

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Config holds the bot configuration.
//
// Example usage:
//
//	config := types.Config{
//	    DiscordToken: "my-token",
//	    PollPeriod:   600,
//	    PollCount:    20,
//	    FreshSeconds: 600,
//	    MsgCount:     10,
//	    ChannelsPath: "/data/channels.txt",
//	    DatabasePath: "/data/stobot.db",
//	    Environment:  "DEV",
//	}
//	if err := config.Validate(); err != nil {
//	    log.Fatal(err)
//	}
type Config struct {
	DiscordToken string // DiscordToken is the token used to authenticate the bot with Discord.
	PollPeriod   int    // PollPeriod is the interval in seconds between polling operations.
	PollCount    int    // PollCount is the number of polls to perform in each operation.
	FreshSeconds int    // FreshSeconds is the threshold in seconds to consider news items fresh.
	MsgCount     int    // MsgCount is the number of messages to process in each operation.
	ChannelsPath string // ChannelsPath is the path to the file containing channel configurations.
	DatabasePath string // DatabasePath is the path to the SQLite database file.
	Environment  string // Environment is the current environment (DEV or PROD) for filtering channels.
}

// Validate checks if the Config is valid. Returns an error if any required field is missing or invalid.
//
// Example:
//
//	err := config.Validate()
//	if err != nil {
//	    // handle error
//	}
func (c *Config) Validate() error {
	if c.DiscordToken == "" {
		return errors.New("discord token is required")
	}
	if c.PollPeriod <= 0 {
		return errors.New("poll period must be positive")
	}
	if c.PollCount <= 0 {
		return errors.New("poll count must be positive")
	}
	if c.FreshSeconds <= 0 {
		return errors.New("fresh seconds must be positive")
	}
	if c.MsgCount <= 0 {
		return errors.New("message count must be positive")
	}
	if c.DatabasePath == "" {
		return errors.New("database path is required")
	}
	if c.Environment != "" && c.Environment != "DEV" && c.Environment != "PROD" {
		return errors.New("environment must be 'DEV' or 'PROD'")
	}
	return nil
}

// Bot represents the Discord bot instance, holding the Discord session, database connection, and configuration.
//
// Example:
//
//	bot := &types.Bot{
//	    Session: dg,
//	    DB:      db,
//	    Config:  &config,
//	}
type Bot struct {
	Session *discordgo.Session // Session is the Discord session used by the bot.
	DB      *sql.DB            // DB is the SQLite database connection used by the bot.
	Config  *Config            // Config is the bot's configuration.
}

// NewsItem represents a news article from the STO API.
//
// Example:
//
//	item := types.NewsItem{
//	    ID: 12345,
//	    Title: "Patch Notes",
//	    Summary: "Details about the latest patch...",
//	    Tags: []string{"patch-notes"},
//	    Platforms: []string{"pc", "xbox"},
//	    Updated: time.Now(),
//	}
type NewsItem struct {
	ID           int64                  `json:"id"`            // ID is the unique identifier of the news item.
	Title        string                 `json:"title"`         // Title is the title of the news item.
	Summary      string                 `json:"summary"`       // Summary is a brief summary of the news item.
	Content      string                 `json:"content"`       // Content is the full content of the news item.
	Tags         []string               `json:"tags"`          // Tags are the tags associated with the news item.
	Platforms    []string               `json:"platforms"`     // Platforms are the platforms associated with the news item.
	Updated      time.Time              `json:"updated"`       // Updated is the timestamp of the last update to the news item.
	ThumbnailURL string                 `json:"thumbnail_url"` // ThumbnailURL is the URL of the thumbnail image for the news item.
	Images       map[string]interface{} `json:"images"`        // Images is a map of image metadata for the news item.
}

// IsEmpty reports whether the NewsItem has no title and no summary.
//
// Example:
//
//	if item.IsEmpty() { /* ... */ }
func (n *NewsItem) IsEmpty() bool {
	return n.Title == "" && n.Summary == ""
}

// HasPlatform reports whether the NewsItem is associated with the given platform (case-insensitive).
//
// Example:
//
//	if item.HasPlatform("pc") { /* ... */ }
func (n *NewsItem) HasPlatform(platform string) bool {
	for _, p := range n.Platforms {
		if strings.EqualFold(p, platform) {
			return true
		}
	}
	return false
}

// HasTag reports whether the NewsItem is associated with the given tag (case-insensitive).
//
// Example:
//
//	if item.HasTag("patch-notes") { /* ... */ }
func (n *NewsItem) HasTag(tag string) bool {
	for _, t := range n.Tags {
		if strings.EqualFold(t, tag) {
			return true
		}
	}
	return false
}

// GetAge returns the time elapsed since the NewsItem was last updated.
//
// Example:
//
//	age := item.GetAge()
func (n *NewsItem) GetAge() time.Duration {
	return time.Since(n.Updated)
}

// String returns a string representation of the NewsItem.
//
// Example:
//
//	fmt.Println(item.String())
func (n *NewsItem) String() string {
	return fmt.Sprintf("NewsItem{ID: %d, Title: %s, Updated: %s, Platforms: %v, Tags: %v}",
		n.ID, n.Title, n.Updated.Format(time.RFC3339), n.Platforms, n.Tags)
}

// UnmarshalJSON implements custom JSON unmarshaling for NewsItem, handling flexible ID and timestamp formats.
func (n *NewsItem) UnmarshalJSON(data []byte) error {
	type Alias NewsItem
	aux := &struct {
		ID      interface{} `json:"id"`      // ID can be a string or a number in the JSON payload.
		Updated string      `json:"updated"` // Updated is the timestamp in string format in the JSON payload.
		*Alias
	}{
		Alias: (*Alias)(n),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Handle ID that might be string or number
	switch v := aux.ID.(type) {
	case string:
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			n.ID = id
		}
	case float64:
		n.ID = int64(v)
	case int64:
		n.ID = v
	}

	// Parse the updated timestamp
	if aux.Updated != "" {
		// Try multiple time formats
		formats := []string{
			time.RFC3339,
			"2006-01-02T15:04:05Z",
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05",
		}

		for _, format := range formats {
			if t, err := time.Parse(format, aux.Updated); err == nil {
				n.Updated = t
				break
			}
		}
	}

	// Extract thumbnail URL from images if available
	if n.Images != nil {
		// Try different thumbnail field names in order of preference
		thumbnailFields := []string{"img_microsite_thumbnail", "thumbnail", "img_microsite_background", "unhighlight_img"}

		for _, field := range thumbnailFields {
			if thumbnail, ok := n.Images[field].(map[string]interface{}); ok {
				if url, ok := thumbnail["url"].(string); ok {
					n.ThumbnailURL = url
					break
				}
			}
		}
	}

	return nil
}

// FetchOptions controls how fetchNews behaves.
//
// Example:
//
//	opts := types.FetchOptions{EnablePagination: true, PageLimit: 5}
type FetchOptions struct {
	EnablePagination bool // EnablePagination determines whether to fetch all pages or stop at the first.
	PageLimit        int  // PageLimit is the maximum number of pages to fetch (0 = unlimited).
	ItemLimit        int  // ItemLimit is the maximum total items to fetch (0 = unlimited).
}

// DatabaseOptions controls how database operations behave.
//
// Example:
//
//	opts := types.DatabaseOptions{UseBatch: true, RetryCount: 5}
type DatabaseOptions struct {
	UseBatch     bool // UseBatch determines whether to use batch operations with transactions.
	IgnoreErrors bool // IgnoreErrors determines whether to continue on individual item errors in batch operations.
	RetryCount   int  // RetryCount is the number of retries on failure (default: 3).
	LogProgress  bool // LogProgress determines whether to log progress for batch operations.
}

// DefaultFetchOptions returns sensible defaults for most fetch operations.
//
// Example:
//
//	opts := types.DefaultFetchOptions()
func DefaultFetchOptions() FetchOptions {
	return FetchOptions{
		EnablePagination: false,
		PageLimit:        0,
		ItemLimit:        0,
	}
}

// DefaultDatabaseOptions returns sensible defaults for regular database operations.
//
// Example:
//
//	opts := types.DefaultDatabaseOptions()
func DefaultDatabaseOptions() DatabaseOptions {
	return DatabaseOptions{
		UseBatch:     false,
		IgnoreErrors: false,
		RetryCount:   3,
		LogProgress:  false,
	}
}

// BatchDatabaseOptions returns optimized settings for batch database operations.
//
// Example:
//
//	opts := types.BatchDatabaseOptions()
func BatchDatabaseOptions() DatabaseOptions {
	return DatabaseOptions{
		UseBatch:     true,
		IgnoreErrors: true,
		RetryCount:   3,
		LogProgress:  true,
	}
}
