// Package database provides persistent storage and migration logic for STOBot.
//
// It manages news caching, channel registration, and posted news tracking using SQLite.
package database

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/FracKenA/sto_news_discord_bot/internal/types"

	log "github.com/sirupsen/logrus"
)

// DatabaseOptions controls how database operations behave
type DatabaseOptions = types.DatabaseOptions

// DefaultDatabaseOptions returns sensible defaults for regular operations
func DefaultDatabaseOptions() DatabaseOptions {
	return types.DefaultDatabaseOptions()
}

// BulkDatabaseOptions returns options optimized for bulk operations
func BulkDatabaseOptions() DatabaseOptions {
	return DatabaseOptions{
		UseBatch:     true,
		IgnoreErrors: true,
		RetryCount:   3,
		LogProgress:  true,
	}
}

// InitDatabase initializes and returns a database connection
func InitDatabase(dbPath string) (*sql.DB, error) {
	return initDatabase(dbPath)
}

func initDatabase(dbPath string) (*sql.DB, error) {
	// Create data directory if it doesn't exist and path starts with /data
	if strings.HasPrefix(dbPath, "/data/") {
		if err := os.MkdirAll("/data", 0755); err != nil {
			return nil, fmt.Errorf("failed to create data directory: %v", err)
		}
	} else {
		// Create parent directory for local database files
		dbDir := filepath.Dir(dbPath)
		if dbDir != "." && dbDir != "" {
			if err := os.MkdirAll(dbDir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create database directory: %v", err)
			}
		}
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	// Create tables
	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %v", err)
	}

	// Add migration to add tags column to existing databases
	if err := migrateDatabase(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %v", err)
	}

	log.Info("Database initialized successfully")
	return db, nil
}

func migrateDatabase(db *sql.DB) error {
	// Check if tags column exists, if not add it
	var tagsColumnExists bool
	err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('news_cache') WHERE name='tags'`).Scan(&tagsColumnExists)
	if err != nil {
		return fmt.Errorf("failed to check for tags column: %v", err)
	}

	if !tagsColumnExists {
		log.Info("Adding tags column to news_cache table")
		if _, err := db.Exec(`ALTER TABLE news_cache ADD COLUMN tags TEXT`); err != nil {
			return fmt.Errorf("failed to add tags column: %v", err)
		}

		// Add index for tags column
		if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_news_cache_tags ON news_cache(tags)`); err != nil {
			return fmt.Errorf("failed to create tags index: %v", err)
		}
	}

	// Check if content column exists, if not add it
	var contentColumnExists bool
	err = db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('news_cache') WHERE name='content'`).Scan(&contentColumnExists)
	if err != nil {
		return fmt.Errorf("failed to check for content column: %v", err)
	}

	if !contentColumnExists {
		log.Info("Adding content column to news_cache table")
		if _, err := db.Exec(`ALTER TABLE news_cache ADD COLUMN content TEXT`); err != nil {
			return fmt.Errorf("failed to add content column: %v", err)
		}
	}

	// Check if old tag column exists (for cleanup)
	var tagColumnExists bool
	err = db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('news_cache') WHERE name='tag'`).Scan(&tagColumnExists)
	if err != nil {
		return fmt.Errorf("failed to check for tag column: %v", err)
	}

	if tagColumnExists {
		log.Info("Found legacy 'tag' column - this can be removed in a future version")
		// Note: We don't automatically remove it to avoid data loss
		// In production, you might want to migrate data from 'tag' to 'tags' first
	}

	// Check if posted_news table has the old schema (news_id as PRIMARY KEY)
	var postedNewsSchema string
	err = db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='posted_news'`).Scan(&postedNewsSchema)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check posted_news schema: %v", err)
	}

	// Check if the schema has the old PRIMARY KEY on news_id
	if strings.Contains(postedNewsSchema, "news_id INTEGER PRIMARY KEY") {
		log.Info("Migrating posted_news table to new schema")

		// Create backup table with old data
		if _, err := db.Exec(`CREATE TABLE posted_news_backup AS SELECT * FROM posted_news`); err != nil {
			return fmt.Errorf("failed to backup posted_news table: %v", err)
		}

		// Drop old table
		if _, err := db.Exec(`DROP TABLE posted_news`); err != nil {
			return fmt.Errorf("failed to drop old posted_news table: %v", err)
		}

		// Recreate with new schema
		if _, err := db.Exec(`CREATE TABLE posted_news (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			news_id INTEGER NOT NULL,
			channel_id TEXT NOT NULL,
			posted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(news_id, channel_id),
			FOREIGN KEY (channel_id) REFERENCES channels(id)
		)`); err != nil {
			return fmt.Errorf("failed to create new posted_news table: %v", err)
		}

		// Restore data from backup
		// Check if posted_at column exists in backup table
		var hasPostedAt bool
		var colCount int
		err = db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('posted_news_backup') WHERE name='posted_at'`).Scan(&colCount)
		if err == nil && colCount > 0 {
			hasPostedAt = true
		}

		if hasPostedAt {
			if _, err := db.Exec(`INSERT OR IGNORE INTO posted_news (news_id, channel_id, posted_at) 
				SELECT news_id, channel_id, posted_at FROM posted_news_backup`); err != nil {
				return fmt.Errorf("failed to restore posted_news data: %v", err)
			}
		} else {
			if _, err := db.Exec(`INSERT OR IGNORE INTO posted_news (news_id, channel_id) 
				SELECT news_id, channel_id FROM posted_news_backup`); err != nil {
				return fmt.Errorf("failed to restore posted_news data: %v", err)
			}
		}

		// Drop backup table
		if _, err := db.Exec(`DROP TABLE posted_news_backup`); err != nil {
			return fmt.Errorf("failed to drop backup table: %v", err)
		}

		// Recreate indexes
		if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_posted_news_channel ON posted_news(channel_id)`); err != nil {
			return fmt.Errorf("failed to create channel index: %v", err)
		}
		if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_posted_news_id ON posted_news(news_id)`); err != nil {
			return fmt.Errorf("failed to create news_id index: %v", err)
		}

		log.Info("Successfully migrated posted_news table")
	}

	// Check if environment column exists in channels table, if not add it
	var environmentColumnExists bool
	err = db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('channels') WHERE name='environment'`).Scan(&environmentColumnExists)
	if err != nil {
		return fmt.Errorf("failed to check for environment column: %v", err)
	}

	if !environmentColumnExists {
		log.Info("Adding environment column to channels table")
		if _, err := db.Exec(`ALTER TABLE channels ADD COLUMN environment TEXT NOT NULL DEFAULT 'PROD' CHECK (environment IN ('DEV', 'PROD'))`); err != nil {
			return fmt.Errorf("failed to add environment column: %v", err)
		}
	}

	return nil
}

func createTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS channels (
			id TEXT PRIMARY KEY,
			platforms TEXT NOT NULL DEFAULT 'pc,xbox,ps',
			environment TEXT NOT NULL DEFAULT 'PROD' CHECK (environment IN ('DEV', 'PROD')),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS posted_news (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			news_id INTEGER NOT NULL,
			channel_id TEXT NOT NULL,
			posted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(news_id, channel_id),
			FOREIGN KEY (channel_id) REFERENCES channels(id)
		)`,
		`CREATE TABLE IF NOT EXISTS news_cache (
			id INTEGER PRIMARY KEY,
			title TEXT NOT NULL,
			summary TEXT,
			content TEXT,
			tags TEXT,
			platforms TEXT,
			updated_at DATETIME,
			thumbnail_url TEXT,
			fetched_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_posted_news_channel ON posted_news(channel_id)`,
		`CREATE INDEX IF NOT EXISTS idx_posted_news_id ON posted_news(news_id)`,
		`CREATE INDEX IF NOT EXISTS idx_news_cache_tags ON news_cache(tags)`,
		`CREATE INDEX IF NOT EXISTS idx_news_cache_updated ON news_cache(updated_at)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %v", err)
		}
	}

	// Add migration to add tags column to existing databases
	if err := migrateDatabase(db); err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	return nil
}

// AddChannel registers a new channel in the database.
func AddChannel(b *types.Bot, channelID string) error {
	// Check if this is a new channel registration
	var exists int
	checkQuery := `SELECT 1 FROM channels WHERE id = ?`
	err := b.DB.QueryRow(checkQuery, channelID).Scan(&exists)
	isNewChannel := (err == sql.ErrNoRows)

	// Register the channel
	query := `INSERT OR REPLACE INTO channels (id, platforms, environment, updated_at) 
			  VALUES (?, 'pc,xbox,ps', 'PROD', CURRENT_TIMESTAMP)`

	_, err = b.DB.Exec(query, channelID)
	if err != nil {
		return fmt.Errorf("failed to add channel: %v", err)
	}

	// If this is a new channel, mark all existing cached news as posted to prevent spam
	if isNewChannel {
		log.Infof("New channel registered: %s, marking existing news as posted", channelID)

		// Get all cached news items
		allNews, err := GetAllCachedNews(b)
		if err != nil {
			log.Errorf("Failed to get cached news for new channel %s: %v", channelID, err)
			// Don't fail the registration, just log the error
		} else if len(allNews) > 0 {
			// Mark all existing news as posted to this new channel using bulk options
			err = MarkMultipleNewsAsPosted(b, allNews, []string{channelID}, BulkDatabaseOptions())
			if err != nil {
				log.Errorf("Failed to mark existing news as posted for new channel %s: %v", channelID, err)
				// Don't fail the registration, just log the error
			} else {
				log.Infof("Marked %d existing news items as posted for new channel %s", len(allNews), channelID)
			}
		}
	}

	return nil
}

// AddChannelWithEnvironment registers a new channel in the database with specified environment.
func AddChannelWithEnvironment(b *types.Bot, channelID string, environment string) error {
	// Validate environment value
	if environment != "DEV" && environment != "PROD" {
		return fmt.Errorf("invalid environment value: %s. Must be 'DEV' or 'PROD'", environment)
	}

	// Check if this is a new channel registration
	var exists int
	checkQuery := `SELECT 1 FROM channels WHERE id = ?`
	err := b.DB.QueryRow(checkQuery, channelID).Scan(&exists)
	isNewChannel := (err == sql.ErrNoRows)

	// Register the channel
	query := `INSERT OR REPLACE INTO channels (id, platforms, environment, updated_at) 
			  VALUES (?, 'pc,xbox,ps', ?, CURRENT_TIMESTAMP)`

	_, err = b.DB.Exec(query, channelID, environment)
	if err != nil {
		return fmt.Errorf("failed to add channel: %v", err)
	}

	// If this is a new channel, mark all existing cached news as posted to prevent spam
	if isNewChannel {
		log.Infof("New channel registered: %s (environment: %s), marking existing news as posted", channelID, environment)

		// Get all cached news items
		allNews, err := GetAllCachedNews(b)
		if err != nil {
			log.Errorf("Failed to get cached news for new channel %s: %v", channelID, err)
			// Don't fail the registration, just log the error
		} else if len(allNews) > 0 {
			// Mark all existing news as posted to this new channel using bulk options
			err = MarkMultipleNewsAsPosted(b, allNews, []string{channelID}, BulkDatabaseOptions())
			if err != nil {
				log.Errorf("Failed to mark existing news as posted for new channel %s: %v", channelID, err)
				// Don't fail the registration, just log the error
			} else {
				log.Infof("Marked %d existing news items as posted for new channel %s", len(allNews), channelID)
			}
		}
	}

	return nil
}

// RemoveChannel removes a channel and its associated posted news entries from the database.
func RemoveChannel(b *types.Bot, channelID string) error {
	tx, err := b.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Printf("Warning: failed to rollback transaction: %v", rollbackErr)
		}
	}()

	// Remove from channels
	_, err = tx.Exec("DELETE FROM channels WHERE id = ?", channelID)
	if err != nil {
		return fmt.Errorf("failed to remove channel: %v", err)
	}

	// Remove posted news entries for this channel
	_, err = tx.Exec("DELETE FROM posted_news WHERE channel_id = ?", channelID)
	if err != nil {
		return fmt.Errorf("failed to remove posted news: %v", err)
	}

	return tx.Commit()
}

// GetChannelPlatforms retrieves the platforms associated with a channel.
func GetChannelPlatforms(b *types.Bot, channelID string) ([]string, error) {
	var platforms string
	query := "SELECT platforms FROM channels WHERE id = ?"

	err := b.DB.QueryRow(query, channelID).Scan(&platforms)
	if err != nil {
		if err == sql.ErrNoRows {
			return []string{}, nil // Channel not registered
		}
		return nil, fmt.Errorf("failed to get channel platforms: %v", err)
	}

	return strings.Split(platforms, ","), nil
}

// GetRegisteredChannels retrieves all registered channel IDs.
func GetRegisteredChannels(b *types.Bot) ([]string, error) {
	query := "SELECT id FROM channels"

	rows, err := b.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query channels: %v", err)
	}
	defer rows.Close()

	var channels []string
	for rows.Next() {
		var channelID string
		if err := rows.Scan(&channelID); err != nil {
			return nil, fmt.Errorf("failed to scan channel: %v", err)
		}
		channels = append(channels, channelID)
	}

	return channels, nil
}

// UpdateChannelPlatforms updates the platforms associated with a channel.
func UpdateChannelPlatforms(b *types.Bot, channelID string, platforms []string) error {
	query := `UPDATE channels SET platforms = ?, updated_at = CURRENT_TIMESTAMP 
			  WHERE id = ?`

	platformsStr := strings.Join(platforms, ",")
	_, err := b.DB.Exec(query, platformsStr, channelID)
	if err != nil {
		return fmt.Errorf("failed to update channel platforms: %v", err)
	}

	return nil
}

// GetChannelEnvironment retrieves the environment associated with a channel.
func GetChannelEnvironment(b *types.Bot, channelID string) (string, error) {
	var environment string
	query := "SELECT environment FROM channels WHERE id = ?"

	err := b.DB.QueryRow(query, channelID).Scan(&environment)
	if err != nil {
		if err == sql.ErrNoRows {
			return "PROD", nil // Default to PROD if channel not found
		}
		return "", fmt.Errorf("failed to get channel environment: %v", err)
	}

	return environment, nil
}

// UpdateChannelEnvironment updates the environment associated with a channel.
func UpdateChannelEnvironment(b *types.Bot, channelID string, environment string) error {
	// Validate environment value
	if environment != "DEV" && environment != "PROD" {
		return fmt.Errorf("invalid environment value: %s. Must be 'DEV' or 'PROD'", environment)
	}

	query := `UPDATE channels SET environment = ?, updated_at = CURRENT_TIMESTAMP 
			  WHERE id = ?`

	result, err := b.DB.Exec(query, environment, channelID)
	if err != nil {
		return fmt.Errorf("failed to update channel environment: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("channel %s not found", channelID)
	}

	return nil
}

// GetChannelsByEnvironment retrieves all channels for a specific environment.
func GetChannelsByEnvironment(b *types.Bot, environment string) ([]string, error) {
	// Validate environment value
	if environment != "DEV" && environment != "PROD" {
		return nil, fmt.Errorf("invalid environment value: %s. Must be 'DEV' or 'PROD'", environment)
	}

	query := "SELECT id FROM channels WHERE environment = ?"

	rows, err := b.DB.Query(query, environment)
	if err != nil {
		return nil, fmt.Errorf("failed to query channels by environment: %v", err)
	}
	defer rows.Close()

	var channels []string
	for rows.Next() {
		var channelID string
		if err := rows.Scan(&channelID); err != nil {
			return nil, fmt.Errorf("failed to scan channel: %v", err)
		}
		channels = append(channels, channelID)
	}

	return channels, nil
}

// IsNewsPosted checks if a news item has been posted to a specific channel.
func IsNewsPosted(b *types.Bot, newsID int64, channelID string) (bool, error) {
	query := "SELECT 1 FROM posted_news WHERE news_id = ? AND channel_id = ?"

	var exists int
	err := b.DB.QueryRow(query, newsID, channelID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if news is posted: %v", err)
	}

	return true, nil
}

// MarkNewsAsPosted marks a news item as posted to a specific channel.
func MarkNewsAsPosted(b *types.Bot, newsID int64, channelID string) error {
	return MarkNewsAsPostedWithOptions(b, newsID, channelID, DefaultDatabaseOptions())
}

// MarkNewsAsPostedWithOptions marks a news item as posted to a specific channel with custom options.
func MarkNewsAsPostedWithOptions(b *types.Bot, newsID int64, channelID string, options DatabaseOptions) error {
	query := `INSERT OR IGNORE INTO posted_news (news_id, channel_id) 
			  VALUES (?, ?)`

	var err error
	for attempt := 0; attempt <= options.RetryCount; attempt++ {
		_, err = b.DB.Exec(query, newsID, channelID)
		if err == nil {
			return nil
		}

		if attempt < options.RetryCount {
			log.Debugf("Retry %d/%d for marking news %d as posted: %v", attempt+1, options.RetryCount, newsID, err)
		}
	}

	return fmt.Errorf("failed to mark news as posted after %d retries: %v", options.RetryCount, err)
}

// MarkMultipleNewsAsPosted marks multiple news items as posted to multiple channels with custom options.
func MarkMultipleNewsAsPosted(b *types.Bot, newsItems []types.NewsItem, channelIDs []string, options DatabaseOptions) error {
	if !options.UseBatch {
		// Single operations
		for _, newsItem := range newsItems {
			for _, channelID := range channelIDs {
				if err := MarkNewsAsPostedWithOptions(b, newsItem.ID, channelID, options); err != nil {
					if !options.IgnoreErrors {
						return err
					}
					log.Debugf("Ignoring error marking news %d as posted to channel %s: %v", newsItem.ID, channelID, err)
				}
			}
		}
		return nil
	}

	// Batch operation with transaction
	tx, err := b.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Printf("Warning: failed to rollback transaction: %v", rollbackErr)
		}
	}()

	query := `INSERT OR IGNORE INTO posted_news (news_id, channel_id) VALUES (?, ?)`

	total := len(newsItems) * len(channelIDs)
	processed := 0

	for _, newsItem := range newsItems {
		for _, channelID := range channelIDs {
			_, err = tx.Exec(query, newsItem.ID, channelID)
			if err != nil {
				if !options.IgnoreErrors {
					return fmt.Errorf("failed to mark news %d as posted to channel %s: %v", newsItem.ID, channelID, err)
				}
				log.Debugf("Ignoring error in batch: news %d to channel %s: %v", newsItem.ID, channelID, err)
			}

			processed++
			if options.LogProgress && processed%100 == 0 {
				log.Infof("Marked %d/%d news items as posted", processed, total)
			}
		}
	}

	if options.LogProgress && processed > 0 {
		log.Infof("Completed marking %d news items as posted", processed)
	}

	return tx.Commit()
}

// CacheNews caches news items in the database.
func CacheNews(b *types.Bot, news []types.NewsItem) error {
	return CacheNewsWithOptions(b, news, DefaultDatabaseOptions())
}

// CacheNewsWithOptions caches news items in the database with custom options.
func CacheNewsWithOptions(b *types.Bot, news []types.NewsItem, options DatabaseOptions) error {
	if len(news) == 0 {
		return nil
	}

	if !options.UseBatch {
		// Single operations
		query := `INSERT OR REPLACE INTO news_cache 
				  (id, title, summary, content, tags, platforms, updated_at, thumbnail_url, fetched_at) 
				  VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`
		for _, item := range news {
			platformsStr := strings.Join(item.Platforms, ",")
			tagsStr := strings.Join(item.Tags, ",")
			var err error
			for attempt := 0; attempt <= options.RetryCount; attempt++ {
				_, err = b.DB.Exec(query, item.ID, item.Title, item.Summary, item.Content,
					tagsStr, platformsStr, item.Updated, item.ThumbnailURL)
				if err == nil {
					break
				}
				if attempt < options.RetryCount {
					log.Debugf("Retry %d/%d for caching news %d: %v", attempt+1, options.RetryCount, item.ID, err)
				}
			}
			// Handle final error after all retries
			if err != nil {
				if !options.IgnoreErrors {
					return fmt.Errorf("failed to cache news item %d after %d retries: %v", item.ID, options.RetryCount, err)
				}
				log.Debugf("Ignoring error caching news item %d: %v", item.ID, err)
			}
		}
		return nil
	}

	// Batch operation with transaction
	tx, err := b.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Printf("Warning: failed to rollback transaction: %v", rollbackErr)
		}
	}()

	query := `INSERT OR REPLACE INTO news_cache 
			  (id, title, summary, content, tags, platforms, updated_at, thumbnail_url, fetched_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`

	for i, item := range news {
		platformsStr := strings.Join(item.Platforms, ",")
		tagsStr := strings.Join(item.Tags, ",")
		_, err = tx.Exec(query, item.ID, item.Title, item.Summary, item.Content,
			tagsStr, platformsStr, item.Updated, item.ThumbnailURL)
		if err != nil {
			if !options.IgnoreErrors {
				return fmt.Errorf("failed to cache news item %d: %v", item.ID, err)
			}
			log.Debugf("Ignoring error in batch caching news item %d: %v", item.ID, err)
		}
		if options.LogProgress && (i+1)%100 == 0 {
			log.Infof("Cached %d/%d news items", i+1, len(news))
		}
	}
	if options.LogProgress && len(news) > 0 {
		log.Infof("Completed caching %d news items", len(news))
	}
	return tx.Commit()
}

// CleanOldCache removes cache entries older than 30 days.
func CleanOldCache(b *types.Bot) error {
	// Remove cache entries older than 30 days
	query := `DELETE FROM news_cache 
			  WHERE fetched_at < datetime('now', '-30 days')`
	result, err := b.DB.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to clean old cache: %v", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Infof("Cleaned %d old cache entries", rowsAffected)
	}
	return nil
}

// ImportChannelsFromFile imports channel configuration from a channels.txt file into the database.
func ImportChannelsFromFile(b *types.Bot, filePath string) error {
	log.Infof("Importing channels from file: %s", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open channels file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	importedCount := 0
	skippedCount := 0

	tx, err := b.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Printf("Warning: failed to rollback transaction: %v", rollbackErr)
		}
	}()

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse channel entry: channel:123456789|pc,ps,xbox
		if !strings.HasPrefix(line, "channel:") {
			log.Warnf("Skipping invalid line: %s", line)
			skippedCount++
			continue
		}

		parts := strings.Split(strings.TrimPrefix(line, "channel:"), "|")
		if len(parts) != 2 {
			log.Warnf("Skipping malformed line: %s", line)
			skippedCount++
			continue
		}

		channelID := strings.TrimSpace(parts[0])
		platformsStr := strings.TrimSpace(parts[1])

		// Validate channel ID is numeric
		if _, err := strconv.ParseUint(channelID, 10, 64); err != nil {
			log.Warnf("Skipping line with invalid channel ID: %s", line)
			skippedCount++
			continue
		}

		// Clean up platforms
		platforms := strings.Split(platformsStr, ",")
		var validPlatforms []string
		for _, platform := range platforms {
			platform = strings.TrimSpace(platform)
			if platform != "" {
				validPlatforms = append(validPlatforms, platform)
			}
		}

		if len(validPlatforms) == 0 {
			validPlatforms = []string{"pc", "xbox", "ps"} // default platforms
		}

		platformsStr = strings.Join(validPlatforms, ",")

		// Check if channel already exists
		var existingPlatforms string
		err := tx.QueryRow("SELECT platforms FROM channels WHERE id = ?", channelID).Scan(&existingPlatforms)
		if err == nil {
			log.Infof("Channel %s already exists with platforms %s, skipping", channelID, existingPlatforms)
			skippedCount++
			continue
		} else if err != sql.ErrNoRows {
			return fmt.Errorf("failed to check existing channel: %v", err)
		}

		// Insert channel
		_, err = tx.Exec(`INSERT INTO channels (id, platforms, environment, created_at, updated_at) 
						  VALUES (?, ?, 'PROD', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
			channelID, platformsStr)
		if err != nil {
			return fmt.Errorf("failed to insert channel %s: %v", channelID, err)
		}

		log.Infof("Imported channel %s with platforms %s", channelID, platformsStr)
		importedCount++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	log.Infof("Import completed: %d channels imported, %d skipped", importedCount, skippedCount)
	return nil
}

// GetAllCachedNews retrieves all cached news items from the database.
func GetAllCachedNews(b *types.Bot) ([]types.NewsItem, error) {
	query := `SELECT id, title, summary, content, tags, platforms, updated_at, thumbnail_url 
			  FROM news_cache 
			  ORDER BY id DESC`

	rows, err := b.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query cached news: %v", err)
	}
	defer rows.Close()

	return parseNewsRows(rows)
}

// SearchNewsContent searches for news items containing the specified text in title, summary, or content.
func SearchNewsContent(b *types.Bot, searchTerm string, limit int) ([]types.NewsItem, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}
	if limit > 25 {
		limit = 25 // Maximum limit to prevent overwhelming Discord
	}

	query := `SELECT id, title, summary, content, tags, platforms, updated_at, thumbnail_url 
			  FROM news_cache 
			  WHERE (title LIKE ? OR summary LIKE ? OR content LIKE ?)
			  AND content IS NOT NULL AND content != ''
			  ORDER BY updated_at DESC
			  LIMIT ?`

	searchPattern := "%" + searchTerm + "%"
	rows, err := b.DB.Query(query, searchPattern, searchPattern, searchPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search news content: %v", err)
	}
	defer rows.Close()

	return parseNewsRows(rows)
}

// SearchNewsByTags searches for news items that contain any of the specified tags.
func SearchNewsByTags(b *types.Bot, tags []string, limit int) ([]types.NewsItem, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 25 {
		limit = 25
	}

	if len(tags) == 0 {
		return []types.NewsItem{}, nil
	}

	// Build WHERE clause for tags
	var conditions []string
	var args []interface{}
	for _, tag := range tags {
		conditions = append(conditions, "tags LIKE ?")
		args = append(args, "%"+tag+"%")
	}

	query := fmt.Sprintf(`SELECT id, title, summary, content, tags, platforms, updated_at, thumbnail_url 
			  FROM news_cache 
			  WHERE (%s)
			  ORDER BY updated_at DESC
			  LIMIT ?`, strings.Join(conditions, " OR "))

	args = append(args, limit)

	rows, err := b.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search by tags: %v", err)
	}
	defer rows.Close()

	return parseNewsRows(rows)
}

// GetRandomNews returns a random news article, optionally filtered by platform.
func GetRandomNews(b *types.Bot, platform string) (*types.NewsItem, error) {
	var query string
	var args []interface{}

	if platform != "" {
		query = `SELECT id, title, summary, content, tags, platforms, updated_at, thumbnail_url 
				 FROM news_cache 
				 WHERE platforms LIKE ?
				 ORDER BY RANDOM() 
				 LIMIT 1`
		args = append(args, "%"+platform+"%")
	} else {
		query = `SELECT id, title, summary, content, tags, platforms, updated_at, thumbnail_url 
				 FROM news_cache 
				 ORDER BY RANDOM() 
				 LIMIT 1`
	}

	rows, err := b.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get random news: %v", err)
	}
	defer rows.Close()

	newsItems, err := parseNewsRows(rows)
	if err != nil {
		return nil, err
	}

	if len(newsItems) == 0 {
		return nil, nil
	}

	return &newsItems[0], nil
}

// GetRecentNews returns recent news items.
func GetRecentNews(b *types.Bot, limit int) ([]types.NewsItem, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	query := `SELECT id, title, summary, content, tags, platforms, updated_at, thumbnail_url 
			  FROM news_cache 
			  ORDER BY updated_at DESC
			  LIMIT ?`

	rows, err := b.DB.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent news: %v", err)
	}
	defer rows.Close()

	return parseNewsRows(rows)
}

// GetDatabaseStats returns statistics about the news database.
func GetDatabaseStats(b *types.Bot) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total news count
	var totalNews int
	err := b.DB.QueryRow("SELECT COUNT(*) FROM news_cache").Scan(&totalNews)
	if err != nil {
		return nil, fmt.Errorf("failed to get total news count: %v", err)
	}
	stats["total_news"] = totalNews

	// Total channels
	var totalChannels int
	err = b.DB.QueryRow("SELECT COUNT(*) FROM channels").Scan(&totalChannels)
	if err != nil {
		return nil, fmt.Errorf("failed to get total channels: %v", err)
	}
	stats["total_channels"] = totalChannels

	// Total posted items
	var totalPosted int
	err = b.DB.QueryRow("SELECT COUNT(*) FROM posted_news").Scan(&totalPosted)
	if err != nil {
		return nil, fmt.Errorf("failed to get total posted count: %v", err)
	}
	stats["total_posted"] = totalPosted

	// Oldest and newest articles
	var oldest, newest sql.NullString
	err = b.DB.QueryRow("SELECT MIN(updated_at), MAX(updated_at) FROM news_cache").Scan(&oldest, &newest)
	if err != nil {
		return nil, fmt.Errorf("failed to get date range: %v", err)
	}

	// Handle NULL values for empty database
	if oldest.Valid && newest.Valid {
		stats["oldest_article"] = oldest.String
		stats["newest_article"] = newest.String
	} else {
		stats["oldest_article"] = ""
		stats["newest_article"] = ""
	}

	return stats, nil
}

// GetPopularTags returns the most frequently used tags.
func GetPopularTags(b *types.Bot, limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 20 {
		limit = 20
	}

	// Get all tags and count them
	rows, err := b.DB.Query("SELECT tags FROM news_cache WHERE tags IS NOT NULL AND tags != ''")
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %v", err)
	}
	defer rows.Close()

	tagCounts := make(map[string]int)
	for rows.Next() {
		var tagsStr string
		if err := rows.Scan(&tagsStr); err != nil {
			continue
		}

		tags := strings.Split(tagsStr, ",")
		for _, tag := range tags {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tagCounts[tag]++
			}
		}
	}

	// Convert to sorted slice
	type tagStat struct {
		Tag   string
		Count int
	}

	var tagStats []tagStat
	for tag, count := range tagCounts {
		tagStats = append(tagStats, tagStat{Tag: tag, Count: count})
	}

	// Sort by count (descending)
	for i := 0; i < len(tagStats)-1; i++ {
		for j := i + 1; j < len(tagStats); j++ {
			if tagStats[j].Count > tagStats[i].Count {
				tagStats[i], tagStats[j] = tagStats[j], tagStats[i]
			}
		}
	}

	// Limit results
	if len(tagStats) > limit {
		tagStats = tagStats[:limit]
	}

	// Convert to return format
	var result []map[string]interface{}
	for _, stat := range tagStats {
		result = append(result, map[string]interface{}{
			"tag":   stat.Tag,
			"count": stat.Count,
		})
	}

	return result, nil
}

// GetTrendingTags returns tags that have appeared frequently in recent news.
func GetTrendingTags(b *types.Bot, days int, limit int) ([]map[string]interface{}, error) {
	if days <= 0 {
		days = 7 // Default to last week
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 20 {
		limit = 20
	}

	cutoffDate := time.Now().AddDate(0, 0, -days)

	rows, err := b.DB.Query(`SELECT tags FROM news_cache 
							 WHERE tags IS NOT NULL AND tags != '' 
							 AND updated_at >= ?`, cutoffDate.Format("2006-01-02 15:04:05"))
	if err != nil {
		return nil, fmt.Errorf("failed to query trending tags: %v", err)
	}
	defer rows.Close()

	tagCounts := make(map[string]int)
	for rows.Next() {
		var tagsStr string
		if err := rows.Scan(&tagsStr); err != nil {
			continue
		}

		tags := strings.Split(tagsStr, ",")
		for _, tag := range tags {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tagCounts[tag]++
			}
		}
	}

	// Convert and sort similar to GetPopularTags
	type tagStat struct {
		Tag   string
		Count int
	}

	var tagStats []tagStat
	for tag, count := range tagCounts {
		tagStats = append(tagStats, tagStat{Tag: tag, Count: count})
	}

	// Sort by count (descending)
	for i := 0; i < len(tagStats)-1; i++ {
		for j := i + 1; j < len(tagStats); j++ {
			if tagStats[j].Count > tagStats[i].Count {
				tagStats[i], tagStats[j] = tagStats[j], tagStats[i]
			}
		}
	}

	// Limit results
	if len(tagStats) > limit {
		tagStats = tagStats[:limit]
	}

	// Convert to return format
	var result []map[string]interface{}
	for _, stat := range tagStats {
		result = append(result, map[string]interface{}{
			"tag":   stat.Tag,
			"count": stat.Count,
		})
	}

	return result, nil
}

// GetChannelEngagement returns engagement statistics for channels.
func GetChannelEngagement(b *types.Bot, channelID string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total posts in this channel
	var totalPosts int
	err := b.DB.QueryRow("SELECT COUNT(*) FROM posted_news WHERE channel_id = ?", channelID).Scan(&totalPosts)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel post count: %v", err)
	}
	stats["total_posts"] = totalPosts

	// Posts in last 7 days
	weekAgo := time.Now().AddDate(0, 0, -7)
	var weeklyPosts int
	err = b.DB.QueryRow(`SELECT COUNT(*) FROM posted_news 
						 WHERE channel_id = ? AND posted_at >= ?`,
		channelID, weekAgo.Format("2006-01-02 15:04:05")).Scan(&weeklyPosts)
	if err != nil {
		return nil, fmt.Errorf("failed to get weekly post count: %v", err)
	}
	stats["weekly_posts"] = weeklyPosts

	// First and last post dates
	var firstPost, lastPost string
	err = b.DB.QueryRow(`SELECT MIN(posted_at), MAX(posted_at) FROM posted_news 
						 WHERE channel_id = ?`, channelID).Scan(&firstPost, &lastPost)
	if err != nil {
		return nil, fmt.Errorf("failed to get post date range: %v", err)
	}
	stats["first_post"] = firstPost
	stats["last_post"] = lastPost

	return stats, nil
}

// GetPopularNewsThisWeek returns the most posted news items from the last week.
func GetPopularNewsThisWeek(b *types.Bot, limit int) ([]types.NewsItem, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 20 {
		limit = 20
	}

	weekAgo := time.Now().AddDate(0, 0, -7)

	query := `SELECT nc.id, nc.title, nc.summary, nc.content, nc.tags, nc.platforms, nc.updated_at, nc.thumbnail_url,
					 COUNT(pn.news_id) as post_count
			  FROM news_cache nc
			  JOIN posted_news pn ON nc.id = pn.news_id
			  WHERE pn.posted_at >= ?
			  GROUP BY nc.id
			  ORDER BY post_count DESC, nc.updated_at DESC
			  LIMIT ?`

	rows, err := b.DB.Query(query, weekAgo.Format("2006-01-02 15:04:05"), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular news: %v", err)
	}
	defer rows.Close()

	var newsItems []types.NewsItem
	for rows.Next() {
		var item types.NewsItem
		var tagsStr, platformsStr string
		var thumbnailURL *string
		var content *string
		var postCount int

		err := rows.Scan(&item.ID, &item.Title, &item.Summary, &content, &tagsStr, &platformsStr, &item.Updated, &thumbnailURL, &postCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan popular news item: %v", err)
		}

		// Parse tags and platforms
		if tagsStr != "" {
			item.Tags = strings.Split(tagsStr, ",")
		}
		if platformsStr != "" {
			item.Platforms = strings.Split(platformsStr, ",")
		}
		if thumbnailURL != nil {
			item.ThumbnailURL = *thumbnailURL
		}
		if content != nil {
			item.Content = *content
		}

		newsItems = append(newsItems, item)
	}

	return newsItems, nil
}

// parseNewsRows is a helper function to parse SQL rows into NewsItem structs.
func parseNewsRows(rows *sql.Rows) ([]types.NewsItem, error) {
	var newsItems []types.NewsItem
	for rows.Next() {
		var item types.NewsItem
		var tagsStr, platformsStr string
		var thumbnailURL *string
		var content *string

		err := rows.Scan(&item.ID, &item.Title, &item.Summary, &content, &tagsStr, &platformsStr, &item.Updated, &thumbnailURL)
		if err != nil {
			return nil, fmt.Errorf("failed to scan news item: %v", err)
		}

		// Parse tags
		if tagsStr != "" {
			item.Tags = strings.Split(tagsStr, ",")
		}

		// Parse platforms
		if platformsStr != "" {
			item.Platforms = strings.Split(platformsStr, ",")
		}

		// Handle thumbnail URL
		if thumbnailURL != nil {
			item.ThumbnailURL = *thumbnailURL
		}

		// Handle content
		if content != nil {
			item.Content = *content
		}

		newsItems = append(newsItems, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading rows: %v", err)
	}

	return newsItems, nil
}

// Convenience functions for testing that wrap the Bot-based functions

// GetChannels gets all registered channel IDs (convenience wrapper)
func GetChannels(db *sql.DB) ([]string, error) {
	bot := &types.Bot{DB: db}
	return GetRegisteredChannels(bot)
}

// StoreNews stores news items (convenience wrapper)
func StoreNews(db *sql.DB, news []types.NewsItem, options DatabaseOptions) error {
	bot := &types.Bot{DB: db}
	return CacheNewsWithOptions(bot, news, options)
}

// GetFreshNews retrieves fresh news items (convenience wrapper)
func GetFreshNews(db *sql.DB, freshSeconds int) ([]types.NewsItem, error) {
	query := `SELECT id, title, summary, content, tags, platforms, updated_at, thumbnail_url
			  FROM news_cache 
			  WHERE updated_at > datetime('now', '-' || ? || ' seconds')
			  ORDER BY updated_at DESC`

	rows, err := db.Query(query, freshSeconds)
	if err != nil {
		return nil, fmt.Errorf("failed to query fresh news: %v", err)
	}
	defer rows.Close()

	return parseNewsRows(rows)
}

// IsPosted checks if news is posted (convenience wrapper)
func IsPosted(db *sql.DB, newsID int64, channelID string) (bool, error) {
	bot := &types.Bot{DB: db}
	return IsNewsPosted(bot, newsID, channelID)
}

// MarkAsPosted marks news as posted (convenience wrapper)
func MarkAsPosted(db *sql.DB, newsID int64, channelID string) error {
	bot := &types.Bot{DB: db}
	return MarkNewsAsPosted(bot, newsID, channelID)
}
