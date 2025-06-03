package discord

import (
	"fmt"
	"strings"

	"github.com/FracKenA/sto_news_discord_bot/internal/database"
	"github.com/FracKenA/sto_news_discord_bot/internal/news"
	"github.com/FracKenA/sto_news_discord_bot/internal/types"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

// handleNews handles the "news" command interaction
func handleNews(b *types.Bot, s *discordgo.Session, i *discordgo.InteractionCreate, tag string) {
	// Check for nil bot
	if b == nil {
		log.Error("Cannot handle news: nil bot provided")
		if s != nil && i != nil {
			Respond(s, i, "❌ Internal error: bot not available.")
		}
		return
	}

	// Acknowledge the interaction first
	Respond(s, i, "🔍 Fetching recent Star Trek Online news...")

	// Get recent news from cache first
	freshNews, err := database.GetFreshNews(b.DB, b.Config.FreshSeconds)
	if err != nil {
		log.Errorf("Failed to get fresh news: %v", err)
		Followup(s, i, "❌ Failed to fetch news. Please try again later.")
		return
	}

	// Filter news by tag if specified
	var filteredNews []types.NewsItem
	if tag != "" && tag != "star-trek-online" {
		for _, newsItem := range freshNews {
			if newsItem.HasTag(tag) {
				filteredNews = append(filteredNews, newsItem)
			}
		}
	} else {
		filteredNews = freshNews
	}

	// If no cached news, try to fetch new news
	if len(filteredNews) == 0 {
		log.Infof("No cached news found, fetching from API for tag: %s", tag)
		newsItems, err := news.FetchNews(b, tag, 5, news.DefaultFetchOptions()) // Fetch 5 recent items
		if err != nil {
			log.Errorf("Failed to fetch news from API: %v", err)
			Followup(s, i, "❌ No recent news found and failed to fetch from API.")
			return
		}
		filteredNews = newsItems
	}

	if len(filteredNews) == 0 {
		Followup(s, i, "📰 No recent news found for the specified criteria.")
		return
	}

	// Limit to 3 most recent items to avoid spam
	if len(filteredNews) > 3 {
		filteredNews = filteredNews[:3]
	}

	// Create a single message with multiple embeds
	var embeds []*discordgo.MessageEmbed
	for _, newsItem := range filteredNews {
		embed := formatNewsEmbed(newsItem)
		embeds = append(embeds, embed)
	}

	// Discord allows a maximum of 10 embeds per message
	const maxEmbedsPerMessage = 10
	for idx := 0; idx < len(embeds); idx += maxEmbedsPerMessage {
		end := idx + maxEmbedsPerMessage
		if end > len(embeds) {
			end = len(embeds)
		}
		content := ""
		if idx == 0 {
			var tagDisplay string
			if tag == "" {
				tagDisplay = "General"
			} else {
				tagDisplay = strings.ToUpper(tag[:1]) + tag[1:]
			}
			content = fmt.Sprintf("📰 **Recent %s News** (%d items)", tagDisplay, len(filteredNews))
		}
		if err := FollowupWithEmbeds(s, i, content, embeds[idx:end]); err != nil {
			log.Errorf("Failed to send news embeds: %v", err)
			if idx == 0 {
				Followup(s, i, "❌ Failed to send news items.")
			}
			return
		}
	}

	log.Infof("Sent %d news items for tag '%s' via slash command", len(filteredNews), tag)
}
