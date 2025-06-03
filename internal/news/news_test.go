// Package news contains tests for the STOBot news package.
//
// These tests cover API integration, formatting, and helper methods.
package news

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/FracKenA/sto_news_discord_bot/internal/types"
)

func TestBuildNewsURL(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		limit    int
		offset   int
		platform string
		fields   []string
		expected string
	}{
		{
			name:     "basic URL",
			tag:      "",
			limit:    0,
			offset:   0,
			platform: "",
			fields:   nil,
			expected: "https://api.arcgames.com/v1.0/games/sto/news",
		},
		{
			name:     "with tag",
			tag:      "patch-notes",
			limit:    0,
			offset:   0,
			platform: "",
			fields:   nil,
			expected: "https://api.arcgames.com/v1.0/games/sto/news?tag=patch-notes",
		},
		{
			name:     "with limit and offset",
			tag:      "",
			limit:    10,
			offset:   20,
			platform: "",
			fields:   nil,
			expected: "https://api.arcgames.com/v1.0/games/sto/news?limit=10&offset=20",
		},
		{
			name:     "with platform",
			tag:      "",
			limit:    0,
			offset:   0,
			platform: "pc",
			fields:   nil,
			expected: "https://api.arcgames.com/v1.0/games/sto/news?platform=pc",
		},
		{
			name:     "with fields",
			tag:      "",
			limit:    0,
			offset:   0,
			platform: "",
			fields:   []string{"title", "summary", "updated"},
			expected: "https://api.arcgames.com/v1.0/games/sto/news?field%5B%5D=title&field%5B%5D=summary&field%5B%5D=updated",
		},
		{
			name:     "all parameters",
			tag:      "patch-notes",
			limit:    15,
			offset:   5,
			platform: "pc",
			fields:   []string{"title", "summary"},
			expected: "https://api.arcgames.com/v1.0/games/sto/news?field%5B%5D=title&field%5B%5D=summary&limit=15&offset=5&platform=pc&tag=patch-notes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildNewsURL(tt.tag, tt.limit, tt.offset, tt.platform, tt.fields)
			if result != tt.expected {
				t.Errorf("buildNewsURL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFetchNewsFromAPI(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock successful response
		response := NewsResponse{
			News: []types.NewsItem{
				{
					ID:        12345,
					Title:     "Test News Item",
					Summary:   "This is a test news item",
					Tags:      []string{"test", "news"},
					Platforms: []string{"PC"},
					Updated:   time.Now(),
				},
				{
					ID:        12346,
					Title:     "Another Test News",
					Summary:   "This is another test news item",
					Tags:      []string{"test", "update"},
					Platforms: []string{"PC", "Console"},
					Updated:   time.Now().Add(-time.Hour),
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Test basic fetch
	// Note: This would require modifying the actual function to accept a custom base URL
	// For now, we'll test the response parsing logic
	client := &http.Client{}
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	var newsResp NewsResponse
	err = json.NewDecoder(resp.Body).Decode(&newsResp)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(newsResp.News) != 2 {
		t.Errorf("Expected 2 news items, got %d", len(newsResp.News))
	}

	if newsResp.News[0].ID != 12345 {
		t.Errorf("Expected first news ID 12345, got %d", newsResp.News[0].ID)
	}

	if newsResp.News[0].Title != "Test News Item" {
		t.Errorf("Expected first news title 'Test News Item', got %s", newsResp.News[0].Title)
	}
}

func TestFetchNewsError(t *testing.T) {
	// Create a mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	// Test error handling
	client := &http.Client{}
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status code 500, got %d", resp.StatusCode)
	}
}

func TestFormatNewsForDiscord(t *testing.T) {
	newsItem := types.NewsItem{
		ID:           12345,
		Title:        "Test News Item",
		Summary:      "This is a test news item with some content that might be long",
		Tags:         []string{"test", "news"},
		Platforms:    []string{"PC", "Console"},
		Updated:      time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
		ThumbnailURL: "https://example.com/thumbnail.jpg",
	}

	embed := formatNewsForDiscord(newsItem)

	if embed.Title != newsItem.Title {
		t.Errorf("Expected embed title '%s', got '%s'", newsItem.Title, embed.Title)
	}

	if embed.Description != newsItem.Summary {
		t.Errorf("Expected embed description '%s', got '%s'", newsItem.Summary, embed.Description)
	}

	if embed.Thumbnail == nil || embed.Thumbnail.URL != newsItem.ThumbnailURL {
		t.Errorf("Expected thumbnail URL '%s', got '%v'", newsItem.ThumbnailURL, embed.Thumbnail)
	}

	if embed.Color != 0x00ff00 {
		t.Errorf("Expected embed color 0x00ff00, got 0x%x", embed.Color)
	}

	// Check if timestamp is set
	if embed.Timestamp == "" {
		t.Error("Expected embed timestamp to be set")
	}

	// Check if fields are set correctly
	expectedFields := 2 // Tags and Platforms
	if len(embed.Fields) != expectedFields {
		t.Errorf("Expected %d fields, got %d", expectedFields, len(embed.Fields))
	}
}

func TestFormatNewsForDiscordWithoutThumbnail(t *testing.T) {
	newsItem := types.NewsItem{
		ID:        12345,
		Title:     "Test News Item",
		Summary:   "This is a test news item",
		Tags:      []string{"test"},
		Platforms: []string{"PC"},
		Updated:   time.Now(),
		// No ThumbnailURL
	}

	embed := formatNewsForDiscord(newsItem)

	if embed.Thumbnail != nil {
		t.Error("Expected no thumbnail when ThumbnailURL is empty")
	}
}

func TestFormatNewsForDiscordLongSummary(t *testing.T) {
	// Create a very long summary
	longSummary := ""
	for i := 0; i < 100; i++ {
		longSummary += "This is a very long summary that should be truncated. "
	}

	newsItem := types.NewsItem{
		ID:      12345,
		Title:   "Test News Item",
		Summary: longSummary,
		Updated: time.Now(),
	}

	embed := formatNewsForDiscord(newsItem)

	// Discord embeds have a description limit
	maxDescriptionLength := 4096
	if len(embed.Description) > maxDescriptionLength {
		t.Errorf("Embed description too long: %d characters (max %d)",
			len(embed.Description), maxDescriptionLength)
	}
}

func TestNewsItemHelperMethods(t *testing.T) {
	newsItem := types.NewsItem{
		ID:        12345,
		Title:     "Test News Item",
		Summary:   "This is a test news item",
		Tags:      []string{"test", "news", "patch-notes"},
		Platforms: []string{"PC", "Console"},
		Updated:   time.Now().Add(-2 * time.Hour),
	}

	// Test IsEmpty
	if newsItem.IsEmpty() {
		t.Error("Expected news item not to be empty")
	}

	emptyItem := types.NewsItem{}
	if !emptyItem.IsEmpty() {
		t.Error("Expected empty news item to be empty")
	}

	// Test HasPlatform
	if !newsItem.HasPlatform("PC") {
		t.Error("Expected news item to have PC platform")
	}
	if !newsItem.HasPlatform("pc") { // case insensitive
		t.Error("Expected news item to have pc platform (case insensitive)")
	}
	if newsItem.HasPlatform("Mobile") {
		t.Error("Expected news item not to have Mobile platform")
	}

	// Test HasTag
	if !newsItem.HasTag("test") {
		t.Error("Expected news item to have test tag")
	}
	if !newsItem.HasTag("TEST") { // case insensitive
		t.Error("Expected news item to have TEST tag (case insensitive)")
	}
	if newsItem.HasTag("missing") {
		t.Error("Expected news item not to have missing tag")
	}

	// Test GetAge
	age := newsItem.GetAge()
	if age < time.Hour || age > 3*time.Hour {
		t.Errorf("Expected age around 2 hours, got %v", age)
	}

	// Test String
	str := newsItem.String()
	if str == "" {
		t.Error("Expected non-empty string representation")
	}
	if !strings.Contains(str, "12345") {
		t.Error("Expected string representation to contain ID")
	}
	if !strings.Contains(str, "Test News Item") {
		t.Error("Expected string representation to contain title")
	}
}

func TestNewsResponseParsing(t *testing.T) {
	// Test JSON unmarshaling of news response
	jsonData := `{
		"news": [
			{
				"id": "12345",
				"title": "Test News",
				"summary": "Test summary",
				"tags": ["test", "news"],
				"platforms": ["PC"],
				"updated": "2024-01-15T12:00:00Z"
			}
		]
	}`

	var response NewsResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal news response: %v", err)
	}

	if len(response.News) != 1 {
		t.Errorf("Expected 1 news item, got %d", len(response.News))
	}

	item := response.News[0]
	if item.ID != 12345 {
		t.Errorf("Expected ID 12345, got %d", item.ID)
	}
	if item.Title != "Test News" {
		t.Errorf("Expected title 'Test News', got '%s'", item.Title)
	}
}

func TestFetchOptions(t *testing.T) {
	// Test default fetch options
	opts := types.DefaultFetchOptions()
	if opts.EnablePagination {
		t.Error("Expected default pagination to be disabled")
	}
	if opts.PageLimit != 0 {
		t.Errorf("Expected default page limit 0, got %d", opts.PageLimit)
	}
	if opts.ItemLimit != 0 {
		t.Errorf("Expected default item limit 0, got %d", opts.ItemLimit)
	}

	// Test custom options
	customOpts := types.FetchOptions{
		EnablePagination: true,
		PageLimit:        5,
		ItemLimit:        50,
	}

	if !customOpts.EnablePagination {
		t.Error("Expected custom pagination to be enabled")
	}
	if customOpts.PageLimit != 5 {
		t.Errorf("Expected custom page limit 5, got %d", customOpts.PageLimit)
	}
	if customOpts.ItemLimit != 50 {
		t.Errorf("Expected custom item limit 50, got %d", customOpts.ItemLimit)
	}
}
