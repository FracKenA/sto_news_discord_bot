package database

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/FracKenA/sto_news_discord_bot/internal/types"
)

// SearchQuery represents a parsed search query with filters
type SearchQuery struct {
	Terms     []string
	Phrases   []string
	MustHave  []string // Required terms (AND)
	MustNot   []string // Excluded terms (NOT)
	Tags      []string
	Platforms []string
	DateFrom  *time.Time
	DateTo    *time.Time
	SortBy    string // "relevance", "date", "title"
	SortOrder string // "asc", "desc"
}

// SearchResult represents a search result with relevance scoring
type SearchResult struct {
	NewsItem types.NewsItem
	Score    float64
	Matches  []string // Which fields matched
}

// ParseSearchQuery parses a complex search query string
func ParseSearchQuery(query string) *SearchQuery {
	sq := &SearchQuery{
		SortBy:    "relevance",
		SortOrder: "desc",
	}

	// Extract quoted phrases first
	phraseRegex := regexp.MustCompile(`"([^"]+)"`)
	phrases := phraseRegex.FindAllStringSubmatch(query, -1)
	for _, phrase := range phrases {
		sq.Phrases = append(sq.Phrases, strings.ToLower(phrase[1]))
	}
	// Remove phrases from query
	query = phraseRegex.ReplaceAllString(query, "")

	// Extract special operators
	tokens := strings.Fields(query)
	for i := 0; i < len(tokens); i++ {
		token := strings.ToLower(tokens[i])

		switch {
		case strings.HasPrefix(token, "+"):
			// Required term: +word
			sq.MustHave = append(sq.MustHave, strings.TrimPrefix(token, "+"))
		case strings.HasPrefix(token, "-"):
			// Excluded term: -word
			sq.MustNot = append(sq.MustNot, strings.TrimPrefix(token, "-"))
		case strings.HasPrefix(token, "tag:"):
			// Tag filter: tag:events
			sq.Tags = append(sq.Tags, strings.TrimPrefix(token, "tag:"))
		case strings.HasPrefix(token, "platform:"):
			// Platform filter: platform:pc
			sq.Platforms = append(sq.Platforms, strings.TrimPrefix(token, "platform:"))
		case strings.HasPrefix(token, "after:"):
			// Date filter: after:2023-01-01
			if date, err := time.Parse("2006-01-02", strings.TrimPrefix(token, "after:")); err == nil {
				sq.DateFrom = &date
			}
		case strings.HasPrefix(token, "before:"):
			// Date filter: before:2023-12-31
			if date, err := time.Parse("2006-01-02", strings.TrimPrefix(token, "before:")); err == nil {
				sq.DateTo = &date
			}
		case strings.HasPrefix(token, "sort:"):
			// Sort order: sort:date
			sq.SortBy = strings.TrimPrefix(token, "sort:")
		case strings.HasPrefix(token, "order:"):
			// Sort direction: order:asc
			sq.SortOrder = strings.TrimPrefix(token, "order:")
		default:
			// Regular search term
			if token != "" {
				sq.Terms = append(sq.Terms, token)
			}
		}
	}

	return sq
}

// AdvancedSearchNews performs advanced search with complex query parsing
func AdvancedSearchNews(b *types.Bot, queryString string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	// Parse the query
	searchQuery := ParseSearchQuery(queryString)

	// Build base SQL query
	var conditions []string
	var args []interface{}

	// Base condition - content must exist
	conditions = append(conditions, "content IS NOT NULL AND content != ''")

	// Add date filters
	if searchQuery.DateFrom != nil {
		conditions = append(conditions, "updated_at >= ?")
		args = append(args, searchQuery.DateFrom.Format("2006-01-02 15:04:05"))
	}
	if searchQuery.DateTo != nil {
		conditions = append(conditions, "updated_at <= ?")
		args = append(args, searchQuery.DateTo.Format("2006-01-02 15:04:05"))
	}

	// Add tag filters
	for _, tag := range searchQuery.Tags {
		conditions = append(conditions, "tags LIKE ?")
		args = append(args, "%"+tag+"%")
	}

	// Add platform filters
	for _, platform := range searchQuery.Platforms {
		conditions = append(conditions, "platforms LIKE ?")
		args = append(args, "%"+platform+"%")
	}

	// Build the main query
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf(`SELECT id, title, summary, content, tags, platforms, updated_at, thumbnail_url 
			  FROM news_cache %s
			  ORDER BY updated_at DESC`, whereClause)

	rows, err := b.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %v", err)
	}
	defer rows.Close()

	newsItems, err := parseNewsRows(rows)
	if err != nil {
		return nil, err
	}

	// Score and filter results
	var results []SearchResult
	for _, item := range newsItems {
		score, matches := scoreNewsItem(item, searchQuery)
		if score > 0 {
			results = append(results, SearchResult{
				NewsItem: item,
				Score:    score,
				Matches:  matches,
			})
		}
	}

	// Sort results
	sortResults(results, searchQuery.SortBy, searchQuery.SortOrder)

	// Limit results
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// scoreNewsItem calculates relevance score for a news item
func scoreNewsItem(item types.NewsItem, query *SearchQuery) (float64, []string) {
	score := 0.0
	var matches []string

	title := strings.ToLower(item.Title)
	summary := strings.ToLower(item.Summary)
	content := strings.ToLower(item.Content)
	allText := title + " " + summary + " " + content

	// Check required terms (must have all)
	for _, term := range query.MustHave {
		if !strings.Contains(allText, term) {
			return 0, nil // Must have all required terms
		}
	}

	// Check excluded terms (must not have any)
	for _, term := range query.MustNot {
		if strings.Contains(allText, term) {
			return 0, nil // Excluded term found
		}
	}

	// Score phrases (highest weight)
	for _, phrase := range query.Phrases {
		if strings.Contains(title, phrase) {
			score += 10.0
			matches = append(matches, "title: \""+phrase+"\"")
		}
		if strings.Contains(summary, phrase) {
			score += 7.0
			matches = append(matches, "summary: \""+phrase+"\"")
		}
		if strings.Contains(content, phrase) {
			score += 5.0
			matches = append(matches, "content: \""+phrase+"\"")
		}
	}

	// Score individual terms
	for _, term := range query.Terms {
		if strings.Contains(title, term) {
			score += 5.0
			matches = append(matches, "title: "+term)
		}
		if strings.Contains(summary, term) {
			score += 3.0
			matches = append(matches, "summary: "+term)
		}
		if strings.Contains(content, term) {
			score += 1.0
			matches = append(matches, "content: "+term)
		}
	}

	// Boost score for recent articles
	now := time.Now()
	if item.Updated.After(now.AddDate(0, 0, -7)) {
		score *= 1.2 // 20% boost for articles from last week
	} else if item.Updated.After(now.AddDate(0, -1, 0)) {
		score *= 1.1 // 10% boost for articles from last month
	}

	return score, matches
}

// sortResults sorts search results based on criteria
func sortResults(results []SearchResult, sortBy, sortOrder string) {
	sort.Slice(results, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "date":
			less = results[i].NewsItem.Updated.Before(results[j].NewsItem.Updated)
		case "title":
			less = strings.ToLower(results[i].NewsItem.Title) < strings.ToLower(results[j].NewsItem.Title)
		default: // "relevance"
			less = results[i].Score < results[j].Score
		}

		if sortOrder == "asc" {
			return less
		}
		return !less
	})
}

// FuzzySearchNews performs fuzzy matching search
func FuzzySearchNews(b *types.Bot, searchTerm string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 25 {
		limit = 25
	}

	// Get all news items
	query := `SELECT id, title, summary, content, tags, platforms, updated_at, thumbnail_url 
			  FROM news_cache 
			  WHERE content IS NOT NULL AND content != ''
			  ORDER BY updated_at DESC
			  LIMIT 500` // Limit to recent items for performance

	rows, err := b.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get news for fuzzy search: %v", err)
	}
	defer rows.Close()

	newsItems, err := parseNewsRows(rows)
	if err != nil {
		return nil, err
	}

	// Calculate fuzzy scores
	var results []SearchResult
	searchTermLower := strings.ToLower(searchTerm)

	for _, item := range newsItems {
		score := calculateFuzzyScore(item, searchTermLower)
		if score > 0.3 { // Minimum threshold
			results = append(results, SearchResult{
				NewsItem: item,
				Score:    score,
				Matches:  []string{"fuzzy match"},
			})
		}
	}

	// Sort by score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Limit results
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// calculateFuzzyScore calculates fuzzy matching score
func calculateFuzzyScore(item types.NewsItem, searchTerm string) float64 {
	title := strings.ToLower(item.Title)
	summary := strings.ToLower(item.Summary)
	content := strings.ToLower(item.Content)

	// Simple fuzzy matching based on substring matching and word overlap
	score := 0.0

	// Exact substring matches
	if strings.Contains(title, searchTerm) {
		score += 1.0
	}
	if strings.Contains(summary, searchTerm) {
		score += 0.7
	}
	if strings.Contains(content, searchTerm) {
		score += 0.5
	}

	// Word-level matching
	searchWords := strings.Fields(searchTerm)
	titleWords := strings.Fields(title)
	summaryWords := strings.Fields(summary)

	for _, searchWord := range searchWords {
		for _, titleWord := range titleWords {
			if strings.Contains(titleWord, searchWord) || strings.Contains(searchWord, titleWord) {
				score += 0.3
			}
		}
		for _, summaryWord := range summaryWords {
			if strings.Contains(summaryWord, searchWord) || strings.Contains(searchWord, summaryWord) {
				score += 0.2
			}
		}
	}

	return score
}

// SearchWithFilters provides a simplified interface for filtered search
func SearchWithFilters(b *types.Bot, options SearchOptions) ([]SearchResult, error) {
	var conditions []string
	var args []interface{}

	// Base condition
	conditions = append(conditions, "content IS NOT NULL AND content != ''")

	// Text search
	if options.Query != "" {
		textCondition := "(title LIKE ? OR summary LIKE ? OR content LIKE ?)"
		conditions = append(conditions, textCondition)
		pattern := "%" + options.Query + "%"
		args = append(args, pattern, pattern, pattern)
	}

	// Tag filter
	if len(options.Tags) > 0 {
		var tagConditions []string
		for _, tag := range options.Tags {
			tagConditions = append(tagConditions, "tags LIKE ?")
			args = append(args, "%"+tag+"%")
		}
		conditions = append(conditions, "("+strings.Join(tagConditions, " OR ")+")")
	}

	// Platform filter
	if len(options.Platforms) > 0 {
		var platformConditions []string
		for _, platform := range options.Platforms {
			platformConditions = append(platformConditions, "platforms LIKE ?")
			args = append(args, "%"+platform+"%")
		}
		conditions = append(conditions, "("+strings.Join(platformConditions, " OR ")+")")
	}

	// Date range
	if options.DateFrom != nil {
		conditions = append(conditions, "updated_at >= ?")
		args = append(args, options.DateFrom.Format("2006-01-02 15:04:05"))
	}
	if options.DateTo != nil {
		conditions = append(conditions, "updated_at <= ?")
		args = append(args, options.DateTo.Format("2006-01-02 15:04:05"))
	}

	// Build and execute query
	whereClause := "WHERE " + strings.Join(conditions, " AND ")
	orderClause := "ORDER BY updated_at DESC"
	if options.SortBy == "title" {
		orderClause = "ORDER BY title"
	}
	if options.SortOrder == "asc" {
		orderClause = strings.Replace(orderClause, "DESC", "ASC", 1)
	}

	query := fmt.Sprintf(`SELECT id, title, summary, content, tags, platforms, updated_at, thumbnail_url 
			  FROM news_cache %s %s LIMIT ?`, whereClause, orderClause)

	limit := options.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	args = append(args, limit)

	rows, err := b.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute filtered search: %v", err)
	}
	defer rows.Close()

	newsItems, err := parseNewsRows(rows)
	if err != nil {
		return nil, err
	}

	// Convert to search results
	var results []SearchResult
	for _, item := range newsItems {
		results = append(results, SearchResult{
			NewsItem: item,
			Score:    1.0, // Default score for filtered results
			Matches:  []string{"filtered search"},
		})
	}

	return results, nil
}

// SearchOptions represents search filter options
type SearchOptions struct {
	Query     string
	Tags      []string
	Platforms []string
	DateFrom  *time.Time
	DateTo    *time.Time
	SortBy    string // "date", "title", "relevance"
	SortOrder string // "asc", "desc"
	Limit     int
}
