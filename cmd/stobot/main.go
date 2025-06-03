// Package main is the entry point for the STOBot application.
//
// STOBot is a Discord bot for Star Trek Online news and channel management.
package main

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/FracKenA/sto_news_discord_bot/internal/database"
	"github.com/FracKenA/sto_news_discord_bot/internal/discord"
	"github.com/FracKenA/sto_news_discord_bot/internal/news"
	"github.com/FracKenA/sto_news_discord_bot/internal/types"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// populateDatabase populates the database with historical news to prevent re-posting old articles.
func populateDatabase(cmd *cobra.Command, args []string) {
	// Get command line flags
	dbPath, _ := cmd.Flags().GetString("database-path")
	count, _ := cmd.Flags().GetInt("count")
	tags, _ := cmd.Flags().GetStringSlice("tags")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Initialize logger
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)

	log.Infof("Populating database with historical news (dry-run: %v)", dryRun)
	log.Infof("Database path: %s", dbPath)
	log.Infof("Count per tag: %d", count)
	log.Infof("Tags: %v", tags)

	// Initialize database
	db, err := database.InitDatabase(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create a minimal bot instance for fetching news
	bot := &types.Bot{
		DB: db,
		Config: &types.Config{
			PollCount: count,
		},
	}

	totalProcessed := 0
	totalCached := 0

	for _, tag := range tags {
		log.Infof("Processing tag: %s", tag)

		// Fetch news for this tag with pagination support
		newsItems, err := news.FetchNews(bot, tag, count, news.BulkFetchOptions())
		if err != nil {
			log.Errorf("Failed to fetch news for tag %s: %v", tag, err)
			continue
		}

		log.Infof("Fetched %d news items for tag: %s", len(newsItems), tag)

		if !dryRun {
			// Cache all news items using bulk options
			if err := news.CacheNewsWithOptions(bot, newsItems, news.BulkDatabaseOptions()); err != nil {
				log.Errorf("Failed to cache news items for tag %s: %v", tag, err)
				continue
			}

			// Get all registered channels to mark news as posted
			channels, err := database.GetRegisteredChannels(bot)
			if err != nil {
				log.Warnf("No registered channels found, skipping posted_news population: %v", err)
			} else {
				// Mark all news as posted to all registered channels using bulk options
				if err := news.MarkMultipleNewsAsPosted(bot, newsItems, channels, news.BulkDatabaseOptions()); err != nil {
					log.Errorf("Failed to mark news items as posted: %v", err)
				} else {
					log.Infof("Marked %d news items as posted to %d channels", len(newsItems), len(channels))
				}
			}
			totalCached += len(newsItems)
		} else {
			log.Infof("DRY RUN: Would cache %d news items for tag %s", len(newsItems), tag)
		}

		totalProcessed += len(newsItems)
	}

	if dryRun {
		log.Infof("DRY RUN COMPLETE: Would have processed %d total news items", totalProcessed)
	} else {
		log.Infof("POPULATE COMPLETE: Processed %d total news items, cached %d items", totalProcessed, totalCached)
	}
}

// importChannels imports channel configuration from a channels.txt file into the database.
func importChannels(cmd *cobra.Command, args []string) {
	// Get command line flags
	dbPath, _ := cmd.Flags().GetString("database-path")
	channelsFile, _ := cmd.Flags().GetString("channels-file")

	// Initialize logger
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)

	log.Infof("Importing channels from %s to database %s", channelsFile, dbPath)

	// Initialize database
	db, err := database.InitDatabase(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create bot instance
	bot := &types.Bot{
		DB: db,
	}

	// Import channels
	err = database.ImportChannelsFromFile(bot, channelsFile)
	if err != nil {
		log.Fatalf("Failed to import channels: %v", err)
	}

	log.Info("Channel import completed successfully")
}

// listChannels lists registered channels in the database.
func listChannels(cmd *cobra.Command, args []string) {
	// Get command line flags
	dbPath, _ := cmd.Flags().GetString("database-path")

	// Initialize logger
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)

	log.Infof("Listing channels from database %s", dbPath)

	// Initialize database
	db, err := database.InitDatabase(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create bot instance
	bot := &types.Bot{
		DB: db,
	}

	// Get registered channels
	channels, err := database.GetRegisteredChannels(bot)
	if err != nil {
		log.Fatalf("Failed to get channels: %v", err)
	}

	if len(channels) == 0 {
		log.Info("No channels registered in database")
		return
	}

	log.Infof("Found %d registered channels:", len(channels))
	for _, channelID := range channels {
		platforms, err := database.GetChannelPlatforms(bot, channelID)
		if err != nil {
			log.Errorf("Failed to get platforms for channel %s: %v", channelID, err)
			continue
		}
		log.Infof("  Channel %s: platforms %v", channelID, platforms)
	}
}

// markAllPosted marks all cached news as already posted to prevent re-sending old messages.
func markAllPosted(cmd *cobra.Command, args []string) {
	// Get command line flags
	dbPath, _ := cmd.Flags().GetString("database-path")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Initialize logger
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)

	log.Infof("Marking all cached news as posted (dry-run: %v)", dryRun)
	log.Infof("Database path: %s", dbPath)

	// Initialize database
	db, err := database.InitDatabase(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create bot instance
	bot := &types.Bot{
		DB: db,
	}

	// Get all registered channels
	channels, err := database.GetRegisteredChannels(bot)
	if err != nil {
		log.Fatalf("Failed to get channels: %v", err)
	}

	if len(channels) == 0 {
		log.Info("No registered channels found")
		return
	}

	log.Infof("Found %d registered channels", len(channels))

	// Get all cached news items
	newsItems, err := database.GetAllCachedNews(bot)
	if err != nil {
		log.Fatalf("Failed to get cached news: %v", err)
	}

	if len(newsItems) == 0 {
		log.Info("No cached news items found")
		return
	}

	log.Infof("Found %d cached news items", len(newsItems))

	if dryRun {
		log.Infof("DRY RUN: Would mark %d news items as posted to %d channels (%d total operations)",
			len(newsItems), len(channels), len(newsItems)*len(channels))
		return
	}

	// Mark all news as posted to all channels
	if err := news.MarkMultipleNewsAsPosted(bot, newsItems, channels, news.BulkDatabaseOptions()); err != nil {
		log.Fatalf("Failed to mark news items as posted: %v", err)
	}

	log.Infof("Successfully marked %d news items as posted to %d channels", len(newsItems), len(channels))
}

// main is the entry point for the STOBot application.
func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Debug("No .env file found or error loading it: ", err)
	}

	var rootCmd = &cobra.Command{
		Use:   "stobot",
		Short: "Star Trek Online Discord News Bot",
		Run:   runBot,
	}

	var config types.Config
	rootCmd.Flags().StringVar(&config.DiscordToken, "token", os.Getenv("DISCORD_TOKEN"), "Discord bot token")
	rootCmd.Flags().IntVar(&config.PollPeriod, "poll-period", getEnvInt("POLL_PERIOD", 600), "Time in seconds between checking for news")
	rootCmd.Flags().IntVar(&config.PollCount, "poll-count", getEnvInt("POLL_COUNT", 20), "Number of news to poll in each period")
	rootCmd.Flags().IntVar(&config.FreshSeconds, "fresh-seconds", getEnvInt("FRESH_SECONDS", 600), "Maximum age of news items to post")
	rootCmd.Flags().IntVar(&config.MsgCount, "msg-count", getEnvInt("MSG_COUNT", 10), "Number of Discord messages to check for duplicates")
	rootCmd.Flags().StringVar(&config.ChannelsPath, "channels-path", getEnvString("CHANNELS_PATH", "/data/channels.txt"), "Path to channels file")
	rootCmd.Flags().StringVar(&config.DatabasePath, "database-path", getEnvString("DATABASE_PATH", "./data/stobot.db"), "Path to SQLite database")

	// Add populate-db subcommand
	var populateCmd = &cobra.Command{
		Use:   "populate-db",
		Short: "Populate database with historical news to prevent re-posting old articles",
		Run:   populateDatabase,
	}
	populateCmd.Flags().StringVar(&config.DatabasePath, "database-path", getEnvString("DATABASE_PATH", "./data/stobot.db"), "Path to SQLite database")
	populateCmd.Flags().IntVar(&config.PollCount, "count", getEnvInt("POLL_COUNT", 100), "Number of news items to fetch and mark as posted")
	populateCmd.Flags().StringSliceP("tags", "t", []string{"star-trek-online", "patch-notes"}, "News tags to populate")
	populateCmd.Flags().BoolP("dry-run", "n", false, "Show what would be populated without making changes")

	// Add import-channels subcommand
	var importCmd = &cobra.Command{
		Use:   "import-channels",
		Short: "Import channel configuration from channels.txt file into database",
		Run:   importChannels,
	}
	importCmd.Flags().StringVar(&config.DatabasePath, "database-path", getEnvString("DATABASE_PATH", "./data/stobot.db"), "Path to SQLite database")
	importCmd.Flags().StringVar(&config.ChannelsPath, "channels-file", getEnvString("CHANNELS_PATH", "./channels.txt"), "Path to channels.txt file to import")

	// Add list-channels subcommand
	var listCmd = &cobra.Command{
		Use:   "list-channels",
		Short: "List registered channels in the database",
		Run:   listChannels,
	}
	listCmd.Flags().StringVar(&config.DatabasePath, "database-path", getEnvString("DATABASE_PATH", "./data/stobot.db"), "Path to SQLite database")

	// Add mark-all-posted subcommand
	var markPostedCmd = &cobra.Command{
		Use:   "mark-all-posted",
		Short: "Mark all cached news as already posted to prevent re-sending old messages",
		Run:   markAllPosted,
	}
	markPostedCmd.Flags().StringVar(&config.DatabasePath, "database-path", getEnvString("DATABASE_PATH", "./data/stobot.db"), "Path to SQLite database")
	markPostedCmd.Flags().BoolP("dry-run", "n", false, "Show what would be marked without making changes")

	rootCmd.AddCommand(populateCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(markPostedCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// runBot initializes and starts the STOBot application.
func runBot(cmd *cobra.Command, args []string) {
	config := &types.Config{}
	config.DiscordToken, _ = cmd.Flags().GetString("token")
	config.PollPeriod, _ = cmd.Flags().GetInt("poll-period")
	config.PollCount, _ = cmd.Flags().GetInt("poll-count")
	config.FreshSeconds, _ = cmd.Flags().GetInt("fresh-seconds")
	config.MsgCount, _ = cmd.Flags().GetInt("msg-count")
	config.ChannelsPath, _ = cmd.Flags().GetString("channels-path")
	config.DatabasePath, _ = cmd.Flags().GetString("database-path")

	if config.DiscordToken == "" {
		log.Fatal("Discord token is required")
	}

	// Initialize logger
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)

	// Initialize database
	db, err := database.InitDatabase(config.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create Discord session
	dg, err := discordgo.New("Bot " + config.DiscordToken)
	if err != nil {
		log.Fatalf("Failed to create Discord session: %v", err)
	}

	bot := &types.Bot{
		Session: dg,
		DB:      db,
		Config:  config,
	}

	// Register event handlers
	dg.AddHandler(discord.Ready(bot))
	dg.AddHandler(discord.InteractionCreate(bot))

	// Set intents
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages

	// Open connection
	err = dg.Open()
	if err != nil {
		log.Fatalf("Failed to open Discord connection: %v", err)
	}
	defer dg.Close()

	log.Info("Bot is now running. Press CTRL-C to exit.")

	// --- CATCH UP ON UNPOSTED NEWS AT STARTUP ---
	go news.CatchUpUnpostedNews(bot, 7) // 7 days catch-up window
	// --------------------------------------------

	// Start news polling
	go news.NewsPoller(bot)

	// Wait for interrupt
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Info("Gracefully shutting down...")
}

// getEnvInt retrieves an integer value from the environment or returns a default value.
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

// getEnvString retrieves a string value from the environment or returns a default value.
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
