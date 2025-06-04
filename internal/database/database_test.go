// Package database contains tests for the STOBot database package.
//
// These tests cover migrations, CRUD operations, and error handling.
package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/FracKenA/sto_news_discord_bot/internal/types"

	_ "github.com/mattn/go-sqlite3"
)

func TestInitDatabase(t *testing.T) {
	// Create temporary directory for test database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	// Test database initialization
	db, err := InitDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Verify database connection
	if err := db.Ping(); err != nil {
		t.Fatalf("Database ping failed: %v", err)
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
			t.Errorf("Table %s does not exist", table)
		}
	}
}

func TestStoreNews(t *testing.T) {
	// Setup test database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := InitDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Test news item
	news := types.NewsItem{
		ID:        12345,
		Title:     "Test News",
		Summary:   "This is a test news item",
		Tags:      []string{"test", "news"},
		Platforms: []string{"PC", "Console"},
		Updated:   time.Now(),
	}

	// Test storing news
	err = StoreNews(db, []types.NewsItem{news}, DefaultDatabaseOptions())
	if err != nil {
		t.Fatalf("Failed to store news: %v", err)
	}

	// Verify news was stored
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM news_cache WHERE id = ?", news.ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query news: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 news item, got %d", count)
	}

	// Test duplicate storage (should not create duplicate)
	err = StoreNews(db, []types.NewsItem{news}, DefaultDatabaseOptions())
	if err != nil {
		t.Fatalf("Failed to store duplicate news: %v", err)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM news_cache WHERE id = ?", news.ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query news after duplicate: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 news item after duplicate storage, got %d", count)
	}
}

func TestGetFreshNews(t *testing.T) {
	// Setup test database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := InitDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Store test news items
	now := time.Now()
	freshNews := types.NewsItem{
		ID:      1,
		Title:   "Fresh News",
		Summary: "This is fresh news",
		Updated: now.Add(-time.Hour), // 1 hour old
	}
	oldNews := types.NewsItem{
		ID:      2,
		Title:   "Old News",
		Summary: "This is old news",
		Updated: now.Add(-48 * time.Hour), // 48 hours old
	}

	err = StoreNews(db, []types.NewsItem{freshNews, oldNews}, DefaultDatabaseOptions())
	if err != nil {
		t.Fatalf("Failed to store test news: %v", err)
	}

	// Test getting fresh news (within 24 hours)
	fresh, err := GetFreshNews(db, 24*60*60) // 24 hours in seconds
	if err != nil {
		t.Fatalf("Failed to get fresh news: %v", err)
	}

	if len(fresh) != 1 {
		t.Errorf("Expected 1 fresh news item, got %d", len(fresh))
	}
	if len(fresh) > 0 && fresh[0].ID != freshNews.ID {
		t.Errorf("Expected fresh news ID %d, got %d", freshNews.ID, fresh[0].ID)
	}
}

func TestGetAllCachedNews(t *testing.T) {
	// Setup test database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := InitDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Store test news items
	newsItems := []types.NewsItem{
		{ID: 1, Title: "News 1", Summary: "Summary 1", Updated: time.Now()},
		{ID: 2, Title: "News 2", Summary: "Summary 2", Updated: time.Now()},
		{ID: 3, Title: "News 3", Summary: "Summary 3", Updated: time.Now()},
	}

	err = StoreNews(db, newsItems, DefaultDatabaseOptions())
	if err != nil {
		t.Fatalf("Failed to store test news: %v", err)
	}

	// Test getting all cached news
	bot := &types.Bot{DB: db}
	cached, err := GetAllCachedNews(bot)
	if err != nil {
		t.Fatalf("Failed to get cached news: %v", err)
	}

	if len(cached) != 3 {
		t.Errorf("Expected 3 cached news items, got %d", len(cached))
	}

	// Verify IDs are present
	ids := make(map[int64]bool)
	for _, item := range cached {
		ids[item.ID] = true
	}
	for _, expected := range newsItems {
		if !ids[expected.ID] {
			t.Errorf("Expected news ID %d not found in cached results", expected.ID)
		}
	}
}

func TestChannelOperations(t *testing.T) {
	// Setup test database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := InitDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Test adding channel
	channelID := "123456789"
	bot := &types.Bot{DB: db}

	err = AddChannel(bot, channelID)
	if err != nil {
		t.Fatalf("Failed to add channel: %v", err)
	}

	// Test getting channels
	channels, err := GetRegisteredChannels(bot)
	if err != nil {
		t.Fatalf("Failed to get channels: %v", err)
	}

	if len(channels) != 1 {
		t.Errorf("Expected 1 channel, got %d", len(channels))
	}
	if len(channels) > 0 && channels[0] != channelID {
		t.Errorf("Expected channel ID %s, got %s", channelID, channels[0])
	}

	// Test adding duplicate channel (should not error)
	err = AddChannel(bot, channelID)
	if err != nil {
		t.Fatalf("Failed to add duplicate channel: %v", err)
	}

	channels, err = GetRegisteredChannels(bot)
	if err != nil {
		t.Fatalf("Failed to get channels after duplicate: %v", err)
	}
	if len(channels) != 1 {
		t.Errorf("Expected 1 channel after duplicate, got %d", len(channels))
	}

	// Test removing channel
	err = RemoveChannel(bot, channelID)
	if err != nil {
		t.Fatalf("Failed to remove channel: %v", err)
	}

	channels, err = GetRegisteredChannels(bot)
	if err != nil {
		t.Fatalf("Failed to get channels after removal: %v", err)
	}
	if len(channels) != 0 {
		t.Errorf("Expected 0 channels after removal, got %d", len(channels))
	}
}

func TestPostingOperations(t *testing.T) {
	// Setup test database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := InitDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Add test channel and news
	channelID := "123456789"
	bot := &types.Bot{DB: db}
	err = AddChannel(bot, channelID)
	if err != nil {
		t.Fatalf("Failed to add channel: %v", err)
	}

	news := types.NewsItem{
		ID:      12345,
		Title:   "Test News",
		Summary: "Test summary",
		Updated: time.Now(),
	}
	err = StoreNews(db, []types.NewsItem{news}, DefaultDatabaseOptions())
	if err != nil {
		t.Fatalf("Failed to store news: %v", err)
	}

	// Test marking as posted
	err = MarkAsPosted(db, news.ID, channelID)
	if err != nil {
		t.Fatalf("Failed to mark as posted: %v", err)
	}

	// Test checking if posted
	posted, err := IsPosted(db, news.ID, channelID)
	if err != nil {
		t.Fatalf("Failed to check if posted: %v", err)
	}
	if !posted {
		t.Error("Expected news to be marked as posted")
	}

	// Test different channel (should be automatically marked as posted due to new channel auto-marking feature)
	otherChannelID := "987654321"
	err = AddChannel(bot, otherChannelID)
	if err != nil {
		t.Fatalf("Failed to add other channel: %v", err)
	}

	posted, err = IsPosted(db, news.ID, otherChannelID)
	if err != nil {
		t.Fatalf("Failed to check if posted to other channel: %v", err)
	}
	if !posted {
		t.Error("Expected news to be automatically marked as posted to new channel")
	}
}

func TestImportChannelsFromFile(t *testing.T) {
	// Setup test database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := InitDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create test channels file
	channelsFile := filepath.Join(tempDir, "channels.txt")
	content := "channel:123456789|pc,xbox,ps\nchannel:987654321|pc,ps\n"
	err = os.WriteFile(channelsFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create channels file: %v", err)
	}

	// Test importing channels
	bot := &types.Bot{DB: db}
	err = ImportChannelsFromFile(bot, channelsFile)
	if err != nil {
		t.Fatalf("Failed to import channels: %v", err)
	}

	// Verify channels were imported - count them
	channels, err := GetRegisteredChannels(bot)
	if err != nil {
		t.Fatalf("Failed to get channels after import: %v", err)
	}
	if len(channels) != 2 {
		t.Errorf("Expected 2 imported channels, got %d", len(channels))
	}

	// Verify channels were imported
	channels, err = GetRegisteredChannels(bot)
	if err != nil {
		t.Fatalf("Failed to get channels after import: %v", err)
	}
	if len(channels) != 2 {
		t.Errorf("Expected 2 channels after import, got %d", len(channels))
	}

	// Test importing again (should skip duplicates - no new channels should be added)
	err = ImportChannelsFromFile(bot, channelsFile)
	if err != nil {
		t.Fatalf("Failed to import channels again: %v", err)
	}

	// Verify channel count is still the same (no duplicates added)
	channels, err = GetRegisteredChannels(bot)
	if err != nil {
		t.Fatalf("Failed to get channels after second import: %v", err)
	}
	if len(channels) != 2 {
		t.Errorf("Expected 2 channels after second import (no duplicates), got %d", len(channels))
	}
}

func TestDatabaseMigration(t *testing.T) {
	// Create a temporary database with old schema
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create old schema
	_, err = db.Exec(`
		CREATE TABLE news (
			id INTEGER PRIMARY KEY,
			title TEXT NOT NULL,
			summary TEXT,
			tags TEXT,
			platforms TEXT,
			updated DATETIME
		);
		CREATE TABLE channels (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL
		);
		CREATE TABLE posted_news (
			news_id INTEGER PRIMARY KEY,
			channel_id TEXT NOT NULL
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create old schema: %v", err)
	}

	// Insert test data
	_, err = db.Exec("INSERT INTO posted_news (news_id, channel_id) VALUES (1, 'channel1')")
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	db.Close()

	// Now initialize with migration
	db, err = InitDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database with migration: %v", err)
	}
	defer db.Close()

	// Verify new schema exists
	var count int
	query := `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='posted_news' 
			  AND sql LIKE '%UNIQUE(news_id, channel_id)%'`
	err = db.QueryRow(query).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check migrated schema: %v", err)
	}
	if count != 1 {
		t.Error("Database migration did not create proper schema")
	}

	// Verify data was preserved
	err = db.QueryRow("SELECT COUNT(*) FROM posted_news WHERE news_id = 1").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check preserved data: %v", err)
	}
	if count != 1 {
		t.Error("Database migration did not preserve data")
	}
}

func TestBatchDatabaseOptions(t *testing.T) {
	opts := BulkDatabaseOptions()

	if !opts.UseBatch {
		t.Error("Expected UseBatch to be true for bulk operations")
	}
	if !opts.IgnoreErrors {
		t.Error("Expected IgnoreErrors to be true for bulk operations")
	}
	if opts.RetryCount != 3 {
		t.Errorf("Expected RetryCount to be 3, got %d", opts.RetryCount)
	}
	if !opts.LogProgress {
		t.Error("Expected LogProgress to be true for bulk operations")
	}
}

func TestSearchNewsContent(t *testing.T) {
	// Setup test database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := InitDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	bot := &types.Bot{DB: db}

	// Create test news items with different content
	testNews := []types.NewsItem{
		{
			ID:        1,
			Title:     "Tholian Assembly Update",
			Summary:   "Updates to Tholian ships and abilities",
			Content:   "The Tholian Assembly has received major updates to their crystalline ship technology. New abilities include enhanced web generators and improved crystal matrix defenses.",
			Tags:      []string{"tholian", "update"},
			Platforms: []string{"pc", "xbox"},
			Updated:   time.Now(),
		},
		{
			ID:        2,
			Title:     "Federation Starship News",
			Summary:   "New starship designs available",
			Content:   "The Federation has unveiled new starship designs featuring advanced quantum torpedo systems and improved warp cores for enhanced deep space exploration.",
			Tags:      []string{"federation", "ships"},
			Platforms: []string{"pc", "ps"},
			Updated:   time.Now(),
		},
		{
			ID:        3,
			Title:     "General Game Update",
			Summary:   "Various improvements",
			Content:   "This update includes bug fixes, performance improvements, and quality of life changes for all players across all platforms.",
			Tags:      []string{"general", "update"},
			Platforms: []string{"pc", "xbox", "ps"},
			Updated:   time.Now(),
		},
	}

	// Store test news
	err = CacheNews(bot, testNews)
	if err != nil {
		t.Fatalf("Failed to cache test news: %v", err)
	}

	// Test search for "Tholian"
	results, err := SearchNewsContent(bot, "Tholian", 10)
	if err != nil {
		t.Fatalf("Failed to search for 'Tholian': %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'Tholian', got %d", len(results))
	}
	if len(results) > 0 && results[0].ID != 1 {
		t.Errorf("Expected news ID 1, got %d", results[0].ID)
	}

	// Test search for "starship"
	results, err = SearchNewsContent(bot, "starship", 10)
	if err != nil {
		t.Fatalf("Failed to search for 'starship': %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'starship', got %d", len(results))
	}
	if len(results) > 0 && results[0].ID != 2 {
		t.Errorf("Expected news ID 2, got %d", results[0].ID)
	}

	// Test search for "update" (should find multiple)
	results, err = SearchNewsContent(bot, "update", 10)
	if err != nil {
		t.Fatalf("Failed to search for 'update': %v", err)
	}
	if len(results) < 2 {
		t.Errorf("Expected at least 2 results for 'update', got %d", len(results))
	}

	// Test search for non-existent term
	results, err = SearchNewsContent(bot, "nonexistent", 10)
	if err != nil {
		t.Fatalf("Failed to search for 'nonexistent': %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results for 'nonexistent', got %d", len(results))
	}

	// Test limit functionality
	results, err = SearchNewsContent(bot, "update", 1)
	if err != nil {
		t.Fatalf("Failed to search with limit 1: %v", err)
	}
	if len(results) > 1 {
		t.Errorf("Expected at most 1 result with limit 1, got %d", len(results))
	}

	// Test case-insensitive search
	results, err = SearchNewsContent(bot, "THOLIAN", 10)
	if err != nil {
		t.Fatalf("Failed to search for 'THOLIAN': %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result for case-insensitive 'THOLIAN', got %d", len(results))
	}
}

func TestChannelEnvironmentOperations(t *testing.T) {
	// Setup test database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := InitDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	bot := &types.Bot{DB: db}
	channelID := "123456789"

	// Test adding channel with default environment
	err = AddChannel(bot, channelID)
	if err != nil {
		t.Fatalf("Failed to add channel: %v", err)
	}

	// Test getting environment (should default to PROD)
	env, err := GetChannelEnvironment(bot, channelID)
	if err != nil {
		t.Fatalf("Failed to get channel environment: %v", err)
	}
	if env != "PROD" {
		t.Errorf("Expected default environment 'PROD', got '%s'", env)
	}

	// Test updating environment to DEV
	err = UpdateChannelEnvironment(bot, channelID, "DEV")
	if err != nil {
		t.Fatalf("Failed to update channel environment: %v", err)
	}

	// Verify environment was updated
	env, err = GetChannelEnvironment(bot, channelID)
	if err != nil {
		t.Fatalf("Failed to get updated channel environment: %v", err)
	}
	if env != "DEV" {
		t.Errorf("Expected environment 'DEV', got '%s'", env)
	}

	// Test updating back to PROD
	err = UpdateChannelEnvironment(bot, channelID, "PROD")
	if err != nil {
		t.Fatalf("Failed to update channel environment back to PROD: %v", err)
	}

	// Verify environment was updated
	env, err = GetChannelEnvironment(bot, channelID)
	if err != nil {
		t.Fatalf("Failed to get channel environment after update: %v", err)
	}
	if env != "PROD" {
		t.Errorf("Expected environment 'PROD', got '%s'", env)
	}

	// Test invalid environment value
	err = UpdateChannelEnvironment(bot, channelID, "INVALID")
	if err == nil {
		t.Error("Expected error for invalid environment value, got nil")
	}

	// Test updating non-existent channel
	err = UpdateChannelEnvironment(bot, "nonexistent", "DEV")
	if err == nil {
		t.Error("Expected error for non-existent channel, got nil")
	}

	// Test adding channel with specific environment
	channelID2 := "987654321"
	err = AddChannelWithEnvironment(bot, channelID2, "DEV")
	if err != nil {
		t.Fatalf("Failed to add channel with environment: %v", err)
	}

	env, err = GetChannelEnvironment(bot, channelID2)
	if err != nil {
		t.Fatalf("Failed to get environment for new channel: %v", err)
	}
	if env != "DEV" {
		t.Errorf("Expected environment 'DEV' for new channel, got '%s'", env)
	}

	// Test getting channels by environment
	prodChannels, err := GetChannelsByEnvironment(bot, "PROD")
	if err != nil {
		t.Fatalf("Failed to get PROD channels: %v", err)
	}
	if len(prodChannels) != 1 || prodChannels[0] != channelID {
		t.Errorf("Expected 1 PROD channel (%s), got %v", channelID, prodChannels)
	}

	devChannels, err := GetChannelsByEnvironment(bot, "DEV")
	if err != nil {
		t.Fatalf("Failed to get DEV channels: %v", err)
	}
	if len(devChannels) != 1 || devChannels[0] != channelID2 {
		t.Errorf("Expected 1 DEV channel (%s), got %v", channelID2, devChannels)
	}

	// Test invalid environment for GetChannelsByEnvironment
	_, err = GetChannelsByEnvironment(bot, "INVALID")
	if err == nil {
		t.Error("Expected error for invalid environment in GetChannelsByEnvironment, got nil")
	}

	// Test invalid environment for AddChannelWithEnvironment
	err = AddChannelWithEnvironment(bot, "999999999", "INVALID")
	if err == nil {
		t.Error("Expected error for invalid environment in AddChannelWithEnvironment, got nil")
	}
}
