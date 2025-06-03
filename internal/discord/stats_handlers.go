package discord

import (
	"fmt"
	"strings"
	"time"

	"github.com/FracKenA/sto_news_discord_bot/internal/database"
	"github.com/FracKenA/sto_news_discord_bot/internal/types"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

// handleNewsStats handles the "news_stats" command interaction
func handleNewsStats(b *types.Bot, s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Acknowledge interaction with timeout handling
	if err := AcknowledgeWithRetry(s, i); err != nil {
		log.Errorf("Failed to acknowledge news_stats command: %v", err)
		return
	}

	// Get database statistics
	log.Info("Getting database statistics")
	stats, err := database.GetDatabaseStats(b)
	if err != nil {
		log.Errorf("Failed to get database stats: %v", err)
		Followup(s, i, "❌ Failed to get database statistics. Please try again later.")
		return
	}

	// Get popular tags
	popularTags, err := database.GetPopularTags(b, 10)
	if err != nil {
		log.Errorf("Failed to get popular tags: %v", err)
		// Continue without popular tags
	}

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title:       "📊 Database Statistics",
		Description: "Overview of cached news and database health",
		Color:       0x0066cc, // Blue color for statistics
		Timestamp:   time.Now().Format("2006-01-02T15:04:05Z"),
	}

	// Add statistics fields
	totalNews := stats["total_news"].(int)
	totalChannels := stats["total_channels"].(int)
	oldestStr := stats["oldest_article"].(string)
	newestStr := stats["newest_article"].(string)

	// Parse dates with null handling
	var dateRangeValue string
	if oldestStr != "" && newestStr != "" {
		// SQLite stores dates with timezone, so use the correct format
		oldest, err := time.Parse("2006-01-02 15:04:05-07:00", oldestStr)
		if err != nil {
			log.Errorf("Failed to parse oldest date '%s': %v", oldestStr, err)
			dateRangeValue = "Invalid date format"
		} else {
			newest, err := time.Parse("2006-01-02 15:04:05-07:00", newestStr)
			if err != nil {
				log.Errorf("Failed to parse newest date '%s': %v", newestStr, err)
				dateRangeValue = "Invalid date format"
			} else {
				dateRangeValue = fmt.Sprintf("%s to %s", oldest.Format("2006-01-02"), newest.Format("2006-01-02"))
			}
		}
	} else {
		dateRangeValue = "No news articles in database"
	}

	embed.Fields = []*discordgo.MessageEmbedField{
		{
			Name:   "📰 Total News Articles",
			Value:  fmt.Sprintf("%d", totalNews),
			Inline: true,
		},
		{
			Name:   "📺 Registered Channels",
			Value:  fmt.Sprintf("%d", totalChannels),
			Inline: true,
		},
		{
			Name:   "📅 Date Range",
			Value:  dateRangeValue,
			Inline: false,
		},
	}

	// Add popular tags if available
	if len(popularTags) > 0 {
		var tagsText strings.Builder
		for i, tagData := range popularTags {
			if i >= 8 { // Limit to top 8 for readability
				break
			}
			tag := tagData["tag"].(string)
			count := tagData["count"].(int)
			tagsText.WriteString(fmt.Sprintf("• **%s** (%d)\n", tag, count))
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "🔝 Most Popular Tags",
			Value: tagsText.String(),
		})
	}

	// Send the result with enhanced error handling
	if err := FollowupWithEmbeds(s, i, "", []*discordgo.MessageEmbed{embed}); err != nil {
		log.Errorf("Failed to send database stats: %v", err)
		Followup(s, i, "❌ Failed to send database statistics.")
		return
	}

	log.Infof("Sent database statistics: %d total news", totalNews)
}

// handleServerStats handles the "server_stats" command interaction
func handleServerStats(b *types.Bot, s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Acknowledge interaction with timeout handling
	if err := AcknowledgeWithRetry(s, i); err != nil {
		log.Errorf("Failed to acknowledge server_stats command: %v", err)
		return
	}

	guildID := i.GuildID
	if guildID == "" {
		Followup(s, i, "❌ This command can only be used in a server.")
		return
	}

	// Get server engagement stats
	log.Infof("Getting server engagement stats for guild: %s", guildID)

	// Get all channels for this guild and aggregate stats
	channels, err := database.GetRegisteredChannels(b)
	if err != nil {
		Followup(s, i, fmt.Sprintf("❌ Failed to get channels: %v", err))
		return
	}

	totalPosts := 0
	weeklyPosts := 0
	activeChannels := 0

	for _, channelID := range channels {
		// Check if this channel belongs to this guild by trying to get channel info
		channel, err := s.Channel(channelID)
		if err != nil || channel.GuildID != guildID {
			continue // Skip channels not in this guild
		}

		activeChannels++
		channelStats, err := database.GetChannelEngagement(b, channelID)
		if err != nil {
			continue // Skip on error
		}

		if posts, ok := channelStats["total_posts"].(int); ok {
			totalPosts += posts
		}
		if weekly, ok := channelStats["weekly_posts"].(int); ok {
			weeklyPosts += weekly
		}
	}

	if totalPosts == 0 {
		Followup(s, i, "📊 No engagement data found for this server.")
		return
	}

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title:       "📊 Server News Engagement",
		Description: "Statistics for this server's news interactions",
		Color:       0x00cc66, // Green color for engagement
		Timestamp:   time.Now().Format("2006-01-02T15:04:05Z"),
	}

	// Add statistics fields
	embed.Fields = []*discordgo.MessageEmbedField{
		{
			Name:   "📝 Total News Posted",
			Value:  fmt.Sprintf("%d", totalPosts),
			Inline: true,
		},
		{
			Name:   "📺 Active Channels",
			Value:  fmt.Sprintf("%d", activeChannels),
			Inline: true,
		},
		{
			Name:   "📈 Posts This Week",
			Value:  fmt.Sprintf("%d", weeklyPosts),
			Inline: true,
		},
	}

	// Send the result with enhanced error handling
	if err := FollowupWithEmbeds(s, i, "", []*discordgo.MessageEmbed{embed}); err != nil {
		log.Errorf("Failed to send server stats: %v", err)
		Followup(s, i, "❌ Failed to send server statistics.")
		return
	}

	log.Infof("Sent server stats for guild: %s", guildID)
}

// handlePopularThisWeek handles the "popular_this_week" command interaction
func handlePopularThisWeek(b *types.Bot, s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Acknowledge interaction with timeout handling
	if err := AcknowledgeWithRetry(s, i); err != nil {
		log.Errorf("Failed to acknowledge popular_this_week command: %v", err)
		return
	}

	// Get popular articles this week
	log.Info("Getting popular articles for this week")
	popularNews, err := database.GetPopularNewsThisWeek(b, 10) // Get top 10
	if err != nil {
		log.Errorf("Failed to get popular news this week: %v", err)
		Followup(s, i, "❌ Failed to get popular articles. Please try again later.")
		return
	}

	if len(popularNews) == 0 {
		Followup(s, i, "📈 No popular articles found for this week.")
		return
	}

	// Format results as embeds
	var embeds []*discordgo.MessageEmbed
	for i, newsItem := range popularNews {
		embed := formatNewsEmbed(newsItem)
		embed.Title = fmt.Sprintf("⭐ #%d - %s", i+1, embed.Title)
		embed.Color = 0xffd700 // Gold color for popular
		embeds = append(embeds, embed)
	}

	// Send results with enhanced error handling
	content := fmt.Sprintf("⭐ **Most Popular Articles This Week** (%d found)", len(popularNews))
	if err := FollowupWithEmbeds(s, i, content, embeds); err != nil {
		log.Errorf("Failed to send popular articles: %v", err)
		Followup(s, i, "❌ Failed to send popular articles.")
		return
	}

	log.Infof("Sent %d popular articles for this week", len(popularNews))
}

// handleTagTrends handles the "tag_trends" command interaction
func handleTagTrends(b *types.Bot, s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Acknowledge interaction with timeout handling
	if err := AcknowledgeWithRetry(s, i); err != nil {
		log.Errorf("Failed to acknowledge tag_trends command: %v", err)
		return
	}

	// Parse command options
	period := "week" // default
	for _, option := range i.ApplicationCommandData().Options {
		if option.Name == "period" {
			period = option.StringValue()
		}
	}

	// Map period to days
	var days int
	var periodName string
	switch period {
	case "week":
		days = 7
		periodName = "Last 7 Days"
	case "month":
		days = 30
		periodName = "Last 30 Days"
	case "quarter":
		days = 90
		periodName = "Last 90 Days"
	default:
		days = 7
		periodName = "Last 7 Days"
	}

	// Get tag trends
	log.Infof("Getting tag trends for %s (%d days)", periodName, days)
	trendingTags, err := database.GetTrendingTags(b, days, 20) // Get top 20
	if err != nil {
		log.Errorf("Failed to get tag trends: %v", err)
		Followup(s, i, "❌ Failed to get tag trends. Please try again later.")
		return
	}

	if len(trendingTags) == 0 {
		Followup(s, i, fmt.Sprintf("📈 No tag trends found for %s.", periodName))
		return
	}

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📈 Tag Trends - %s", periodName),
		Description: "Most frequently appearing tags in news articles",
		Color:       0xff6600, // Orange color for trends
		Timestamp:   time.Now().Format("2006-01-02T15:04:05Z"),
	}

	// Format trending tags
	var trendsText strings.Builder
	for i, tagData := range trendingTags {
		if i >= 15 { // Limit to top 15 for readability
			break
		}
		tag := tagData["tag"].(string)
		count := tagData["count"].(int)
		trendsText.WriteString(fmt.Sprintf("%d. **%s** (%d)\n", i+1, tag, count))
	}

	embed.Description = trendsText.String()

	// Send the result with enhanced error handling
	if err := FollowupWithEmbeds(s, i, "", []*discordgo.MessageEmbed{embed}); err != nil {
		log.Errorf("Failed to send tag trends: %v", err)
		Followup(s, i, "❌ Failed to send tag trends.")
		return
	}

	log.Infof("Sent tag trends for %s", periodName)
}

// handleEngagementReport handles the "engagement_report" command interaction
func handleEngagementReport(b *types.Bot, s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Validate inputs
	if i == nil || i.Interaction == nil {
		log.Warning("handleEngagementReport called with nil interaction")
		return
	}

	// Check if user is an administrator
	if !hasAdminPermission(s, i) {
		Respond(s, i, "❌ This command requires Administrator permissions.")
		return
	}

	// Acknowledge interaction with timeout handling
	if err := AcknowledgeWithRetry(s, i); err != nil {
		log.Errorf("Failed to acknowledge engagement_report command: %v", err)
		return
	}

	// Get engagement report by aggregating various stats
	log.Info("Getting detailed engagement report")

	// Get database stats for report context
	_, err := database.GetDatabaseStats(b)
	if err != nil {
		log.Errorf("Failed to get database stats: %v", err)
		Followup(s, i, "❌ Failed to get engagement report. Please try again later.")
		return
	}

	channels, err := database.GetRegisteredChannels(b)
	if err != nil {
		log.Errorf("Failed to get registered channels: %v", err)
		Followup(s, i, "❌ Failed to get engagement report. Please try again later.")
		return
	}

	// Calculate engagement metrics
	totalServers := 0 // We'll need to implement guild counting
	totalChannels := len(channels)
	totalPosts := 0
	weeklyPosts := 0

	// Aggregate channel engagement
	for _, channelID := range channels {
		channelStats, err := database.GetChannelEngagement(b, channelID)
		if err != nil {
			continue // Skip on error
		}

		if posts, ok := channelStats["total_posts"].(int); ok {
			totalPosts += posts
		}
		if weekly, ok := channelStats["weekly_posts"].(int); ok {
			weeklyPosts += weekly
		}
	}

	// Calculate daily average
	dailyAverage := float64(weeklyPosts) / 7.0

	// Create detailed embed
	embed := &discordgo.MessageEmbed{
		Title:       "📈 Detailed Engagement Report",
		Description: "Comprehensive bot usage and engagement statistics",
		Color:       0x9932cc, // Purple color for reports
		Timestamp:   time.Now().Format("2006-01-02T15:04:05Z"),
	}

	embed.Fields = []*discordgo.MessageEmbedField{
		{
			Name:   "🏢 Total Servers",
			Value:  fmt.Sprintf("%d", totalServers),
			Inline: true,
		},
		{
			Name:   "📺 Total Channels",
			Value:  fmt.Sprintf("%d", totalChannels),
			Inline: true,
		},
		{
			Name:   "📝 Total Posts",
			Value:  fmt.Sprintf("%d", totalPosts),
			Inline: true,
		},
		{
			Name:   "📈 Weekly Posts",
			Value:  fmt.Sprintf("%d", weeklyPosts),
			Inline: true,
		},
		{
			Name:   "📊 Daily Average",
			Value:  fmt.Sprintf("%.1f", dailyAverage),
			Inline: true,
		},
	}

	// Send the result with enhanced error handling
	if err := FollowupWithEmbeds(s, i, "", []*discordgo.MessageEmbed{embed}); err != nil {
		log.Errorf("Failed to send engagement report: %v", err)
		Followup(s, i, "❌ Failed to send engagement report.")
		return
	}

	log.Info("Sent detailed engagement report")
}
