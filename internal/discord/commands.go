// Package discord provides Discord integration utilities for STOBot.
//
// It includes command registration, event handlers, and Discord API helpers.
package discord

import (
	"github.com/FracKenA/sto_news_discord_bot/internal/types"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

// RegisterCommands registers all slash commands with Discord
func RegisterCommands(s *discordgo.Session) {
	// Wait for the session to be ready and get application info
	if s.State == nil || s.State.User == nil {
		log.Error("Session state is not ready, cannot register commands")
		return
	}

	// For bot applications, the application ID is typically the bot's user ID
	appID := s.State.User.ID
	log.Infof("Registering commands for application ID: %s", appID)

	// First, get existing commands to clean up any obsolete ones
	existingCommands, err := s.ApplicationCommands(appID, "")
	if err != nil {
		log.Warnf("Failed to get existing commands: %v", err)
	} else {
		log.Infof("Found %d existing commands", len(existingCommands))
	}

	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "stobot_register",
			Description: "Register this channel for STO news updates",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "platforms",
					Description: "Comma-separated list of platforms (pc,xbox,ps)",
					Required:    false,
				},
			},
		},
		{
			Name:        "stobot_unregister",
			Description: "Unregister this channel from STO news updates",
		},
		{
			Name:        "stobot_status",
			Description: "Show bot status and registered channels",
		},
		{
			Name:        "stobot_news",
			Description: "Get recent Star Trek Online news",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "tag",
					Description: "News category",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "General", Value: "star-trek-online"},
						{Name: "Patch Notes", Value: "patch-notes"},
						{Name: "Events", Value: "events"},
						{Name: "Dev Blogs", Value: "dev-blogs"},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "platforms",
					Description: "Comma-separated list of platforms (pc,xbox,ps)",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "weeks",
					Description: "Number of weeks back to search (default: 1)",
					Required:    false,
				},
			},
		},
		{
			Name:        "stobot_news_stats",
			Description: "Show database statistics and popular topics",
		},
		{
			Name:        "stobot_server_stats",
			Description: "Show this server's news engagement statistics",
		},
		{
			Name:        "stobot_popular_this_week",
			Description: "Show most engaged articles this week",
		},
		{
			Name:        "stobot_tag_trends",
			Description: "Show trending tags over time",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "period",
					Description: "Time period to analyze",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Last 7 days", Value: "week"},
						{Name: "Last 30 days", Value: "month"},
						{Name: "Last 90 days", Value: "quarter"},
					},
				},
			},
		},
		{
			Name:        "stobot_engagement_report",
			Description: "Show detailed engagement statistics (Admin only)",
		},
		{
			Name:        "stobot_help",
			Description: "Show help information",
		},
		{
			Name:        "stobot_game_status",
			Description: "Check Star Trek Online server status",
		},
		{
			Name:        "stobot_advanced_search",
			Description: "Advanced search with operators and filters",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "query",
					Description: "Advanced search query (use quotes, +required, -excluded, tag:, platform:, etc.)",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "limit",
					Description: "Number of results to return (1-25, default: 10)",
					Required:    false,
				},
			},
		},
		{
			Name:        "stobot_fuzzy_search",
			Description: "Find similar articles using fuzzy matching",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "query",
					Description: "Search term for fuzzy matching",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "limit",
					Description: "Number of results to return (1-25, default: 10)",
					Required:    false,
				},
			},
		},
		{
			Name:        "stobot_filtered_search",
			Description: "Search with multiple filters and sorting options",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "query",
					Description: "Text to search for (optional)",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "tags",
					Description: "Tags to filter by (comma-separated)",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "platforms",
					Description: "Platforms to filter by (comma-separated: pc,xbox,ps)",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "after",
					Description: "Show articles after this date (YYYY-MM-DD)",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "before",
					Description: "Show articles before this date (YYYY-MM-DD)",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "sort",
					Description: "Sort by field",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Date", Value: "date"},
						{Name: "Title", Value: "title"},
						{Name: "Relevance", Value: "relevance"},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "order",
					Description: "Sort order",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Descending", Value: "desc"},
						{Name: "Ascending", Value: "asc"},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "limit",
					Description: "Number of results to return (1-50, default: 10)",
					Required:    false,
				},
			},
		},
	}

	log.Infof("Starting to register %d commands...", len(commands))

	// Create a map of current command names for comparison
	currentCommandNames := make(map[string]bool)
	for _, cmd := range commands {
		currentCommandNames[cmd.Name] = true
	}

	// Remove commands that are no longer in our current list
	for _, existingCmd := range existingCommands {
		if !currentCommandNames[existingCmd.Name] {
			log.Infof("Removing obsolete command: %s", existingCmd.Name)
			err := s.ApplicationCommandDelete(appID, "", existingCmd.ID)
			if err != nil {
				log.Warnf("Failed to delete obsolete command %s: %v", existingCmd.Name, err)
			} else {
				log.Infof("Successfully removed obsolete command: %s", existingCmd.Name)
			}
		}
	}

	successCount := 0
	for i, command := range commands {
		log.Infof("Registering command %d/%d: %s", i+1, len(commands), command.Name)

		// Register as global commands using the application ID
		createdCmd, err := s.ApplicationCommandCreate(appID, "", command)
		if err != nil {
			log.Errorf("Failed to register command %s: %v", command.Name, err)
			// Continue registering other commands even if one fails
		} else {
			log.Infof("Successfully registered command: %s (ID: %s)", command.Name, createdCmd.ID)
			successCount++
		}
	}

	log.Infof("Command registration completed: %d/%d commands registered successfully", successCount, len(commands))
}

// HandleCommand routes slash command interactions to their handlers
func HandleCommand(b *types.Bot, s *discordgo.Session, i *discordgo.InteractionCreate) {
	if b == nil || s == nil || i == nil || i.Interaction == nil {
		log.Warn("HandleCommand called with nil parameters")
		return
	}

	data := i.ApplicationCommandData()
	switch data.Name {
	case "stobot_register":
		handleRegister(b, s, i)
	case "stobot_unregister":
		handleUnregister(b, s, i)
	case "stobot_status":
		handleStatus(b, s, i)
	case "stobot_news":
		tag := "star-trek-online" // default
		if len(data.Options) > 0 {
			for _, option := range data.Options {
				if option.Name == "tag" && option.StringValue() != "" {
					tag = option.StringValue()
					break
				}
			}
		}
		handleNews(b, s, i, tag)
	case "stobot_news_stats":
		handleNewsStats(b, s, i)
	case "stobot_server_stats":
		handleServerStats(b, s, i)
	case "stobot_popular_this_week":
		handlePopularThisWeek(b, s, i)
	case "stobot_tag_trends":
		handleTagTrends(b, s, i)
	case "stobot_engagement_report":
		handleEngagementReport(b, s, i)
	case "stobot_help":
		handleHelp(b, s, i)
	case "stobot_game_status":
		handleGameStatus(b, s, i)
	case "stobot_advanced_search":
		handleAdvancedSearchNews(b, s, i)
	case "stobot_fuzzy_search":
		handleFuzzySearchNews(b, s, i)
	case "stobot_filtered_search":
		handleFilteredSearch(b, s, i)
	}
}

// handleHelp handles the "help" command interaction
func handleHelp(b *types.Bot, s *discordgo.Session, i *discordgo.InteractionCreate) {
	helpText := "**Star Trek Online News Bot**\n\n" +
		"**📰 Basic Commands:**\n" +
		"• `/stobot_news [tag] [platforms] [weeks]` - Get recent STO news\n" +
		"• `/stobot_status` - Show bot status and settings\n" +
		"• `/stobot_game_status` - Check Star Trek Online server status\n\n" +
		"**🔍 Search & Discovery:**\n" +
		"• `/stobot_advanced_search <query> [limit]` - Advanced search with operators\n" +
		"• `/stobot_fuzzy_search <query> [limit]` - Find similar articles\n" +
		"• `/stobot_filtered_search [options]` - Search with filters and sorting\n\n" +
		"**🔍 Advanced Search Syntax:**\n" +
		"• **Phrases:** \"exact phrase\" (use quotes)\n" +
		"• **Required:** +word (must contain)\n" +
		"• **Excluded:** -word (must not contain)\n" +
		"• **Tags:** tag:events, tag:patch-notes\n" +
		"• **Platforms:** platform:pc, platform:xbox\n" +
		"• **Date filters:** after:2023-01-01, before:2023-12-31\n\n" +
		"**📊 Analytics & Stats:**\n" +
		"• `/stobot_news_stats` - Database statistics\n" +
		"• `/stobot_server_stats` - Server engagement stats\n" +
		"• `/stobot_popular_this_week` - Most engaged articles\n" +
		"• `/stobot_tag_trends [period]` - Trending tags over time\n\n" +
		"**⚙️ Admin Commands:**\n" +
		"• `/stobot_register [platforms]` - Register this channel for STO news updates\n" +
		"• `/stobot_unregister` - Unregister this channel from news updates\n" +
		"• `/stobot_engagement_report` - Detailed usage statistics (Admin only)\n\n" +
		"**Platforms:** pc, xbox, ps (comma-separated)\n" +
		"**News Tags:** star-trek-online, patch-notes, events, dev-blogs\n\n" +
		"The bot automatically posts new STO news to registered channels."

	Respond(s, i, helpText)
}
