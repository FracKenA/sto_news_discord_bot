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

// handleAdvancedSearchNews handles the "advanced_search" command interaction
func handleAdvancedSearchNews(b *types.Bot, s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Acknowledge interaction
	if err := AcknowledgeWithRetry(s, i); err != nil {
		log.Errorf("Failed to acknowledge advanced_search command: %v", err)
		return
	}

	// Parse command options
	var query string
	limit := 10

	for _, option := range i.ApplicationCommandData().Options {
		switch option.Name {
		case "query":
			query = option.StringValue()
		case "limit":
			if option.IntValue() > 0 && option.IntValue() <= 25 {
				limit = int(option.IntValue())
			}
		}
	}

	if query == "" {
		Followup(s, i, "❌ Search query is required.")
		return
	}

	// Perform advanced search
	log.Infof("Performing advanced search for: %s (limit: %d)", query, limit)
	results, err := database.AdvancedSearchNews(b, query, limit)
	if err != nil {
		log.Errorf("Failed to perform advanced search: %v", err)
		Followup(s, i, "❌ Failed to perform advanced search. Please try again later.")
		return
	}

	if len(results) == 0 {
		helpText := buildSearchHelpText()
		Followup(s, i, fmt.Sprintf("🔍 No news articles found matching \"%s\".\n\n%s", query, helpText))
		return
	}

	// Format results as embeds
	var embeds []*discordgo.MessageEmbed
	for i, result := range results {
		embed := formatAdvancedSearchResultEmbed(result, i+1)
		embeds = append(embeds, embed)
	}

	// Send results
	content := fmt.Sprintf("🔍 **Advanced search results for \"%s\"** (%d found)", query, len(results))
	if err := FollowupWithEmbeds(s, i, content, embeds); err != nil {
		log.Errorf("Failed to send advanced search results: %v", err)
		Followup(s, i, "❌ Failed to send search results.")
		return
	}

	log.Infof("Sent %d advanced search results", len(results))
}

// handleFuzzySearchNews handles the "fuzzy_search" command interaction
func handleFuzzySearchNews(b *types.Bot, s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Acknowledge interaction
	if err := AcknowledgeWithRetry(s, i); err != nil {
		log.Errorf("Failed to acknowledge fuzzy_search command: %v", err)
		return
	}

	// Parse command options
	var query string
	limit := 10

	for _, option := range i.ApplicationCommandData().Options {
		switch option.Name {
		case "query":
			query = option.StringValue()
		case "limit":
			if option.IntValue() > 0 && option.IntValue() <= 25 {
				limit = int(option.IntValue())
			}
		}
	}

	if query == "" {
		Followup(s, i, "❌ Search query is required.")
		return
	}

	// Perform fuzzy search
	log.Infof("Performing fuzzy search for: %s (limit: %d)", query, limit)
	results, err := database.FuzzySearchNews(b, query, limit)
	if err != nil {
		log.Errorf("Failed to perform fuzzy search: %v", err)
		Followup(s, i, "❌ Failed to perform fuzzy search. Please try again later.")
		return
	}

	if len(results) == 0 {
		Followup(s, i, fmt.Sprintf("🔍 No similar articles found for \"%s\".", query))
		return
	}

	// Format results as embeds
	var embeds []*discordgo.MessageEmbed
	for i, result := range results {
		embed := formatFuzzySearchResultEmbed(result, i+1)
		embeds = append(embeds, embed)
	}

	// Send results
	content := fmt.Sprintf("🔍 **Fuzzy search results for \"%s\"** (%d found)", query, len(results))
	if err := FollowupWithEmbeds(s, i, content, embeds); err != nil {
		log.Errorf("Failed to send fuzzy search results: %v", err)
		Followup(s, i, "❌ Failed to send search results.")
		return
	}

	log.Infof("Sent %d fuzzy search results", len(results))
}

// handleFilteredSearch handles the "filtered_search" command interaction
func handleFilteredSearch(b *types.Bot, s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Acknowledge interaction
	if err := AcknowledgeWithRetry(s, i); err != nil {
		log.Errorf("Failed to acknowledge filtered_search command: %v", err)
		return
	}

	// Parse command options
	options := database.SearchOptions{
		Limit:     10,
		SortBy:    "date",
		SortOrder: "desc",
	}

	for _, option := range i.ApplicationCommandData().Options {
		switch option.Name {
		case "query":
			options.Query = option.StringValue()
		case "tags":
			tagStr := option.StringValue()
			if tagStr != "" {
				options.Tags = strings.Split(strings.ReplaceAll(tagStr, " ", ""), ",")
			}
		case "platforms":
			platformStr := option.StringValue()
			if platformStr != "" {
				options.Platforms = strings.Split(strings.ReplaceAll(platformStr, " ", ""), ",")
			}
		case "after":
			if date, err := time.Parse("2006-01-02", option.StringValue()); err == nil {
				options.DateFrom = &date
			}
		case "before":
			if date, err := time.Parse("2006-01-02", option.StringValue()); err == nil {
				options.DateTo = &date
			}
		case "sort":
			options.SortBy = option.StringValue()
		case "order":
			options.SortOrder = option.StringValue()
		case "limit":
			if option.IntValue() > 0 && option.IntValue() <= 50 {
				options.Limit = int(option.IntValue())
			}
		}
	}

	// Perform filtered search
	log.Infof("Performing filtered search with options: %+v", options)
	results, err := database.SearchWithFilters(b, options)
	if err != nil {
		log.Errorf("Failed to perform filtered search: %v", err)
		Followup(s, i, "❌ Failed to perform filtered search. Please try again later.")
		return
	}

	if len(results) == 0 {
		Followup(s, i, "🔍 No articles found matching the specified filters.")
		return
	}

	// Format results as embeds
	var embeds []*discordgo.MessageEmbed
	for i, result := range results {
		embed := formatFilteredSearchResultEmbed(result, i+1)
		embeds = append(embeds, embed)
	}

	// Send results
	var queryDesc strings.Builder
	if options.Query != "" {
		queryDesc.WriteString(fmt.Sprintf("Query: \"%s\"", options.Query))
	}
	if len(options.Tags) > 0 {
		if queryDesc.Len() > 0 {
			queryDesc.WriteString(", ")
		}
		queryDesc.WriteString(fmt.Sprintf("Tags: %s", strings.Join(options.Tags, ", ")))
	}
	if len(options.Platforms) > 0 {
		if queryDesc.Len() > 0 {
			queryDesc.WriteString(", ")
		}
		queryDesc.WriteString(fmt.Sprintf("Platforms: %s", strings.Join(options.Platforms, ", ")))
	}

	content := fmt.Sprintf("🔍 **Filtered search results** (%d found)\n**Filters:** %s", len(results), queryDesc.String())
	if err := FollowupWithEmbeds(s, i, content, embeds); err != nil {
		log.Errorf("Failed to send filtered search results: %v", err)
		Followup(s, i, "❌ Failed to send search results.")
		return
	}

	log.Infof("Sent %d filtered search results", len(results))
}

// formatAdvancedSearchResultEmbed formats a search result with relevance score
func formatAdvancedSearchResultEmbed(result database.SearchResult, rank int) *discordgo.MessageEmbed {
	embed := formatNewsEmbed(result.NewsItem)

	// Add rank and score information
	embed.Title = fmt.Sprintf("#%d - %s", rank, embed.Title)
	embed.Color = 0x9932cc // Purple for advanced search

	// Add relevance information
	if len(result.Matches) > 0 {
		matchesText := strings.Join(result.Matches[:min(3, len(result.Matches))], ", ")
		if len(result.Matches) > 3 {
			matchesText += fmt.Sprintf(" (+%d more)", len(result.Matches)-3)
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🎯 Relevance",
			Value:  fmt.Sprintf("Score: %.1f\nMatches: %s", result.Score, matchesText),
			Inline: false,
		})
	}

	return embed
}

// formatFuzzySearchResultEmbed formats a fuzzy search result
func formatFuzzySearchResultEmbed(result database.SearchResult, rank int) *discordgo.MessageEmbed {
	embed := formatNewsEmbed(result.NewsItem)

	// Add rank information
	embed.Title = fmt.Sprintf("#%d - %s", rank, embed.Title)
	embed.Color = 0x00ced1 // Dark turquoise for fuzzy search

	// Add similarity score
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "🔍 Similarity",
		Value:  fmt.Sprintf("%.1f%%", result.Score*100),
		Inline: true,
	})

	return embed
}

// formatFilteredSearchResultEmbed formats a filtered search result
func formatFilteredSearchResultEmbed(result database.SearchResult, rank int) *discordgo.MessageEmbed {
	embed := formatNewsEmbed(result.NewsItem)

	// Add rank information
	embed.Title = fmt.Sprintf("#%d - %s", rank, embed.Title)
	embed.Color = 0x32cd32 // Lime green for filtered search

	return embed
}

// buildSearchHelpText provides help text for advanced search syntax
func buildSearchHelpText() string {
	return `**🔍 Advanced Search Syntax:**
• **Phrases:** "exact phrase" (use quotes)
• **Required:** +word (must contain)
• **Excluded:** -word (must not contain)
• **Tags:** tag:events, tag:patch-notes
• **Platforms:** platform:pc, platform:xbox
• **Date filters:** after:2023-01-01, before:2023-12-31
• **Sorting:** sort:date, sort:title, order:asc

**Examples:**
• ` + "`" + `"patch notes" +update -maintenance` + "`" + `
• ` + "`" + `tag:events platform:pc after:2023-01-01` + "`" + `
• ` + "`" + `starship +new sort:date order:asc` + "`" + ``
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
