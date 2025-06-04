// Package news provides Star Trek Online news fetching, parsing, and formatting utilities.
//
// It includes API integration, Discord formatting, and news item helpers.
package news

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/FracKenA/sto_news_discord_bot/internal/database"
	"github.com/FracKenA/sto_news_discord_bot/internal/types"

	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

// NewsResponse is a local struct for API responses
type NewsResponse struct {
	News []types.NewsItem `json:"news"`
}

// buildNewsURL constructs the Arc Games API URL for STO news
func buildNewsURL(tag string, limit int, offset int, platform string, fields []string) string {
	baseURL := "https://api.arcgames.com/v1.0/games/sto/news"
	params := url.Values{}

	if tag != "" {
		params.Add("tag", tag)
	}
	if limit > 0 {
		params.Add("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		params.Add("offset", fmt.Sprintf("%d", offset))
	}
	for _, field := range fields {
		params.Add("field[]", field)
	}
	if platform != "" {
		params.Add("platform", platform)
	}

	if len(params) > 0 {
		return baseURL + "?" + params.Encode()
	}
	return baseURL
}

// FetchOptions controls how news fetching behaves
type FetchOptions struct {
	EnablePagination bool // Whether to use pagination for large requests
	ItemLimit        int  // Items per page when using pagination (default: 100)
}

// DefaultFetchOptions returns sensible defaults for regular bot operation
func DefaultFetchOptions() types.FetchOptions {
	return types.FetchOptions{
		EnablePagination: false,
		ItemLimit:        100,
	}
}

// BulkFetchOptions returns options optimized for bulk operations
func BulkFetchOptions() types.FetchOptions {
	return types.FetchOptions{
		EnablePagination: true,
		ItemLimit:        100,
	}
}

// CacheNewsWithOptions caches news items in the database with specific options.
func CacheNewsWithOptions(b *types.Bot, newsItems []types.NewsItem, options types.DatabaseOptions) error {
	return database.CacheNewsWithOptions(b, newsItems, options)
}

// BulkDatabaseOptions returns database options optimized for bulk operations.
func BulkDatabaseOptions() types.DatabaseOptions {
	return database.BulkDatabaseOptions()
}

// MarkMultipleNewsAsPosted marks multiple news items as posted in the database for specific channels.
func MarkMultipleNewsAsPosted(b *types.Bot, newsItems []types.NewsItem, channels []string, options types.DatabaseOptions) error {
	return database.MarkMultipleNewsAsPosted(b, newsItems, channels, options)
}

// NewsPoller periodically polls for news and processes them for registered channels.
func NewsPoller(b *types.Bot) {
	ticker := time.NewTicker(time.Duration(b.Config.PollPeriod) * time.Second)
	defer ticker.Stop()

	log.Info("News poller started")

	for range ticker.C {
		// Only get channels that match the current environment
		var channels []string
		var err error
		if b.Config.Environment != "" {
			channels, err = database.GetChannelsByEnvironment(b, b.Config.Environment)
			if err != nil {
				log.Errorf("Failed to get channels for environment %s: %v", b.Config.Environment, err)
				continue
			}
		} else {
			// If no environment is set, get all channels (backwards compatibility)
			channels, err = database.GetRegisteredChannels(b)
			if err != nil {
				log.Errorf("Failed to get registered channels: %v", err)
				continue
			}
		}

		if len(channels) == 0 {
			log.Debug("No registered channels found")
			continue
		}

		for _, channelID := range channels {
			go ProcessChannelNews(b, channelID)
		}

		// Clean old cache every poll cycle
		if err := database.CleanOldCache(b); err != nil {
			log.Errorf("Failed to clean old cache: %v", err)
		}
	}
}

// FetchNews fetches news items with pagination and options.
func FetchNews(b *types.Bot, tag string, count int, options types.FetchOptions) ([]types.NewsItem, error) {
	fields := []string{"id", "title", "summary", "tags", "platforms", "updated", "images", "content"}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Determine if we should use pagination
	if !options.EnablePagination || count <= options.ItemLimit {
		// Single request for small counts or when pagination is disabled
		url := buildNewsURL(tag, count, 0, "", fields)
		log.Debugf("Fetching news from: %s", url)

		resp, err := client.Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch news: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
		}

		var newsResponse NewsResponse
		if err := json.NewDecoder(resp.Body).Decode(&newsResponse); err != nil {
			return nil, fmt.Errorf("failed to decode news response: %v", err)
		}

		// Process tags for all items
		processNewsItemTags(newsResponse.News, tag)

		// Clean HTML content for all items
		cleanNewsItemContent(newsResponse.News)

		log.Infof("Fetched %d news items with tag '%s'", len(newsResponse.News), tag)
		return newsResponse.News, nil
	}

	// Use pagination for large requests
	var allNews []types.NewsItem
	offset := 0
	itemLimit := options.ItemLimit

	for len(allNews) < count {
		// Calculate how many items to request in this batch
		remaining := count - len(allNews)
		limit := itemLimit
		if remaining < itemLimit {
			limit = remaining
		}

		url := buildNewsURL(tag, limit, offset, "", fields)
		log.Debugf("Fetching news page: offset=%d, limit=%d, url=%s", offset, limit, url)

		resp, err := client.Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch news page at offset %d: %v", offset, err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("API returned status %d at offset %d", resp.StatusCode, offset)
		}

		var newsResponse NewsResponse
		if err := json.NewDecoder(resp.Body).Decode(&newsResponse); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode news response at offset %d: %v", offset, err)
		}
		resp.Body.Close()

		// Process tags for all items
		processNewsItemTags(newsResponse.News, tag)

		// Clean HTML content for all items
		cleanNewsItemContent(newsResponse.News)

		allNews = append(allNews, newsResponse.News...)
		log.Infof("Fetched page with %d news items (total: %d/%d)", len(newsResponse.News), len(allNews), count)

		// Check if there are more pages
		if len(newsResponse.News) == 0 {
			log.Infof("No more news available for tag '%s'", tag)
			break
		}

		offset += len(newsResponse.News)
	}

	log.Infof("Fetched %d total news items with tag '%s'", len(allNews), tag)
	return allNews, nil
}

// processNewsItemTags ensures the requested tag is included in the tags array.
func processNewsItemTags(newsItems []types.NewsItem, requestedTag string) {
	for i := range newsItems {
		// Ensure the requested tag is in the tags array if it's not already there
		tagExists := false
		for _, existingTag := range newsItems[i].Tags {
			if existingTag == requestedTag {
				tagExists = true
				break
			}
		}
		if !tagExists && requestedTag != "" {
			newsItems[i].Tags = append(newsItems[i].Tags, requestedTag)
		}
	}
}

// cleanNewsItemContent cleans HTML content from news items for better searchability.
func cleanNewsItemContent(newsItems []types.NewsItem) {
	for i := range newsItems {
		if newsItems[i].Content != "" {
			newsItems[i].Content = extractTextFromHTML(newsItems[i].Content)
		}
	}
}

// filterNewsByPlatforms filters news items by the specified platforms.
func filterNewsByPlatforms(news []types.NewsItem, platforms []string) []types.NewsItem {
	if len(platforms) == 0 {
		return news
	}

	platformSet := make(map[string]bool)
	for _, platform := range platforms {
		platformSet[strings.ToLower(platform)] = true
	}

	var filtered []types.NewsItem
	for _, item := range news {
		for _, itemPlatform := range item.Platforms {
			if platformSet[strings.ToLower(itemPlatform)] {
				filtered = append(filtered, item)
				break
			}
		}
	}

	return filtered
}

// IsNewsFresh checks if a news item is fresh.
func IsNewsFresh(b *types.Bot, newsItem types.NewsItem) bool {
	freshThreshold := time.Duration(b.Config.FreshSeconds) * time.Second
	return time.Since(newsItem.Updated) <= freshThreshold
}

// ProcessChannelNews processes news for a channel.
func ProcessChannelNews(b *types.Bot, channelID string) {
	// Check if this channel matches the bot's environment
	if b.Config.Environment != "" {
		channelEnv, err := database.GetChannelEnvironment(b, channelID)
		if err != nil {
			log.Errorf("Failed to get environment for channel %s: %v", channelID, err)
			return
		}
		if channelEnv != b.Config.Environment {
			log.Debugf("Skipping channel %s (environment %s, bot environment %s)", channelID, channelEnv, b.Config.Environment)
			return
		}
	}

	platforms, err := database.GetChannelPlatforms(b, channelID)
	if err != nil {
		log.Errorf("Failed to get platforms for channel %s: %v", channelID, err)
		return
	}
	if len(platforms) == 0 {
		log.Debugf("Channel %s not registered", channelID)
		return
	}

	// Fetch all news at once (no tag or platform filtering)
	newsItems, err := FetchNews(b, "", b.Config.PollCount, DefaultFetchOptions())
	if err != nil {
		log.Errorf("Failed to fetch news: %v", err)
		return
	}

	// Write all news to DB (cache)
	if err := database.CacheNews(b, newsItems); err != nil {
		log.Errorf("Failed to cache news items: %v", err)
	}

	// Post all unposted news
	for _, newsItem := range newsItems {
		posted, err := database.IsNewsPosted(b, newsItem.ID, channelID)
		if err != nil {
			log.Errorf("Failed to check if news %d is posted: %v", newsItem.ID, err)
			continue
		}
		if posted {
			continue
		}
		if err := PostNewsToChannel(b, channelID, newsItem); err != nil {
			log.Errorf("Failed to post news %d to channel %s: %v", newsItem.ID, channelID, err)
			continue
		}
		if err := database.MarkNewsAsPosted(b, newsItem.ID, channelID); err != nil {
			log.Errorf("Failed to mark news %d as posted: %v", newsItem.ID, err)
		}
		log.Infof("Posted news item %d ('%s') to channel %s", newsItem.ID, newsItem.Title, channelID)
	}
}

// IsDuplicateInRecentMessages checks for duplicate news in recent messages.
func IsDuplicateInRecentMessages(b *types.Bot, channelID string, newsItem types.NewsItem) bool {
	messages, err := b.Session.ChannelMessages(channelID, b.Config.MsgCount, "", "", "")
	if err != nil {
		if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "Missing Access") {
			log.Warnf("[IsDuplicateInRecentMessages] Missing access to read messages in channel %s. Skipping duplicate check.", channelID)
			return false // Don't block posting if we can't check
		}
		log.Errorf("Failed to get recent messages for channel %s: %v", channelID, err)
		return false
	}

	// Create a simple title matcher
	titleWords := strings.Fields(strings.ToLower(newsItem.Title))
	if len(titleWords) == 0 {
		return false
	}

	for _, message := range messages {
		if message.Author.ID != b.Session.State.User.ID {
			continue // Only check our own messages
		}

		messageText := strings.ToLower(message.Content)

		// Check embeds too
		for _, embed := range message.Embeds {
			if embed.Title != "" {
				messageText += " " + strings.ToLower(embed.Title)
			}
			if embed.Description != "" {
				messageText += " " + strings.ToLower(embed.Description)
			}
		}

		// Simple word matching - if most title words appear in the message, consider it a duplicate
		matchCount := 0
		for _, word := range titleWords {
			if len(word) > 3 && strings.Contains(messageText, word) {
				matchCount++
			}
		}

		// If more than half the significant words match, consider it a duplicate
		if matchCount > len(titleWords)/2 && matchCount >= 2 {
			return true
		}
	}

	return false
}

// formatNewsForDiscord creates a Discord embed for a news item.
func formatNewsForDiscord(newsItem types.NewsItem) *discordgo.MessageEmbed {
	// Truncate summary to fit Discord's embed description limit
	summary := newsItem.Summary
	if len(summary) > 2048 {
		if len(summary) <= 3 {
			summary = summary[:2048]
		} else {
			summary = summary[:2045] + "..."
		}
	}

	embed := &discordgo.MessageEmbed{
		Title:       newsItem.Title,
		Description: summary,
		URL:         fmt.Sprintf("https://playstartrekonline.com/en/news/article/%d", newsItem.ID),
		Color:       0x00ff00, // Green color
		Timestamp:   newsItem.Updated.Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Platforms: %s", strings.Join(newsItem.Platforms, ", ")),
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Tags",
				Value:  strings.Join(newsItem.Tags, ", "),
				Inline: true,
			},
			{
				Name:   "Platforms",
				Value:  strings.Join(newsItem.Platforms, ", "),
				Inline: true,
			},
		},
	}

	if newsItem.ThumbnailURL != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: newsItem.ThumbnailURL,
		}
	}

	return embed
}

// PostNewsToChannel posts a news item to a Discord channel.
func PostNewsToChannel(b *types.Bot, channelID string, newsItem types.NewsItem) error {
	embed := formatNewsForDiscord(newsItem)
	_, err := b.Session.ChannelMessageSendEmbed(channelID, embed)
	return err
}

// extractTextFromHTML extracts plain text from HTML content, removing all tags and cleaning whitespace.
func extractTextFromHTML(htmlContent string) string {
	if htmlContent == "" {
		return ""
	}

	// Parse HTML content
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		// If parsing fails, fall back to regex-based tag removal
		return cleanHTMLWithRegex(htmlContent)
	}

	// Remove script and style elements completely
	doc.Find("script, style, iframe, img, video, audio").Remove()

	// Extract text content
	text := doc.Text()

	// Clean up whitespace
	return cleanWhitespace(text)
}

// cleanHTMLWithRegex removes HTML tags using regex as a fallback.
func cleanHTMLWithRegex(htmlContent string) string {
	// Remove HTML tags
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(htmlContent, " ")

	// Decode common HTML entities
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")
	text = strings.ReplaceAll(text, "&rsquo;", "'")
	text = strings.ReplaceAll(text, "&lsquo;", "'")
	text = strings.ReplaceAll(text, "&rdquo;", "\"")
	text = strings.ReplaceAll(text, "&ldquo;", "\"")

	return cleanWhitespace(text)
}

// cleanWhitespace normalizes whitespace in text content.
func cleanWhitespace(text string) string {
	// Replace multiple whitespace characters with single spaces
	re := regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")

	// Trim leading and trailing whitespace
	return strings.TrimSpace(text)
}
