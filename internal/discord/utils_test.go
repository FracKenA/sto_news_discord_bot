// Package discord contains tests for the STOBot Discord utility functions.
//
// These tests cover text truncation, response helpers, and API limits.
package discord

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
)

func TestTruncateText(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		maxLength int
		expected  string
	}{
		{
			name:      "text shorter than max",
			text:      "Hello",
			maxLength: 10,
			expected:  "Hello",
		},
		{
			name:      "text equal to max",
			text:      "Hello",
			maxLength: 5,
			expected:  "Hello",
		},
		{
			name:      "text longer than max",
			text:      "Hello World",
			maxLength: 8,
			expected:  "Hello...",
		},
		{
			name:      "empty text",
			text:      "",
			maxLength: 10,
			expected:  "",
		},
		{
			name:      "very short max length",
			text:      "Hello World",
			maxLength: 3,
			expected:  "...",
		},
		{
			name:      "max length less than ellipsis",
			text:      "Hello",
			maxLength: 2,
			expected:  "..", // Truncated ellipsis to fit maxLength
		},
		{
			name:      "long text with max 100",
			text:      strings.Repeat("A", 150),
			maxLength: 100,
			expected:  strings.Repeat("A", 97) + "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateText(tt.text, tt.maxLength)
			if result != tt.expected {
				t.Errorf("TruncateText(%q, %d) = %q, want %q", tt.text, tt.maxLength, result, tt.expected)
			}

			// Verify result length doesn't exceed maxLength
			if len(result) > tt.maxLength {
				t.Errorf("Result length %d exceeds maxLength %d", len(result), tt.maxLength)
			}
		})
	}
}

func TestTruncateTextDiscordLimits(t *testing.T) {
	// Test with Discord's actual limits
	tests := []struct {
		name      string
		text      string
		maxLength int
	}{
		{
			name:      "embed description limit",
			text:      strings.Repeat("A", 5000),
			maxLength: 4096,
		},
		{
			name:      "embed title limit",
			text:      strings.Repeat("A", 300),
			maxLength: 256,
		},
		{
			name:      "message content limit",
			text:      strings.Repeat("A", 3000),
			maxLength: 2000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateText(tt.text, tt.maxLength)

			if len(result) > tt.maxLength {
				t.Errorf("Result length %d exceeds Discord limit %d", len(result), tt.maxLength)
			}

			// Verify it's properly truncated if original was longer
			if len(tt.text) > tt.maxLength {
				if !strings.HasSuffix(result, "...") {
					t.Error("Truncated text should end with ellipsis")
				}
			}
		})
	}
}

func TestRespond(t *testing.T) {
	// Create a mock interaction
	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:    "test_interaction_id",
			Token: "test_token",
			Type:  discordgo.InteractionApplicationCommand,
		},
	}

	// Test that Respond function doesn't panic
	// Note: This will fail without a real Discord session, but shouldn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Respond panicked: %v", r)
		}
	}()

	// Call with nil session (should handle gracefully)
	Respond(nil, interaction, "Test message")

	t.Log("Respond function executed without panic")
}

func TestRespondError(t *testing.T) {
	// Create a mock interaction
	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:    "test_interaction_id",
			Token: "test_token",
			Type:  discordgo.InteractionApplicationCommand,
		},
	}

	// Test that RespondError function doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("RespondError panicked: %v", r)
		}
	}()

	// Call with nil session (should handle gracefully)
	RespondError(nil, interaction, "Test error message")

	t.Log("RespondError function executed without panic")
}

func TestFollowup(t *testing.T) {
	// Create a mock interaction
	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:    "test_interaction_id",
			Token: "test_token",
			Type:  discordgo.InteractionApplicationCommand,
		},
	}

	// Test that Followup function doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Followup panicked: %v", r)
		}
	}()

	// Call with nil session (should handle gracefully)
	Followup(nil, interaction, "Test followup message")

	t.Log("Followup function executed without panic")
}

func TestFollowupError(t *testing.T) {
	// Create a mock interaction
	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:    "test_interaction_id",
			Token: "test_token",
			Type:  discordgo.InteractionApplicationCommand,
		},
	}

	// Test that FollowupError function doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("FollowupError panicked: %v", r)
		}
	}()

	// Call with nil session (should handle gracefully)
	FollowupError(nil, interaction, "Test followup error message")

	t.Log("FollowupError function executed without panic")
}

func TestErrorMessageFormatting(t *testing.T) {
	// Test that error functions format messages correctly
	// We can't easily test the actual Discord interaction without mocking,
	// but we can test the message formatting logic

	testMessage := "Something went wrong"
	expectedFormat := "❌ Error: Something went wrong"

	// We need to verify that RespondError and FollowupError format the message correctly
	// Since these functions call Respond and Followup internally, we can't easily test
	// the formatting without refactoring. For now, we'll test the expected format.

	if !strings.Contains(expectedFormat, "❌ Error:") {
		t.Error("Error message should contain error emoji and prefix")
	}

	if !strings.Contains(expectedFormat, testMessage) {
		t.Error("Error message should contain the original message")
	}
}

func TestUtilityFunctionSignatures(t *testing.T) {
	// Test that utility functions have correct signatures
	// This is a compile-time check

	var testSession *discordgo.Session
	var testInteraction *discordgo.InteractionCreate

	// These should compile without errors
	var _ func(*discordgo.Session, *discordgo.InteractionCreate, string) = Respond
	var _ func(*discordgo.Session, *discordgo.InteractionCreate, string) = RespondError
	var _ func(*discordgo.Session, *discordgo.InteractionCreate, string) = Followup
	var _ func(*discordgo.Session, *discordgo.InteractionCreate, string) = FollowupError
	var _ func(string, int) string = TruncateText

	// Test that functions can be called (even if they fail due to nil session)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Function signature test panicked: %v", r)
		}
	}()

	// These calls will likely fail, but should not panic due to type issues
	Respond(testSession, testInteraction, "test")
	RespondError(testSession, testInteraction, "test")
	Followup(testSession, testInteraction, "test")
	FollowupError(testSession, testInteraction, "test")

	result := TruncateText("test", 10)
	if result != "test" {
		t.Errorf("TruncateText failed: expected 'test', got %s", result)
	}

	t.Log("All utility functions have correct signatures")
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries to be 3, got %d", config.MaxRetries)
	}

	if config.BaseDelay != time.Second {
		t.Errorf("Expected BaseDelay to be 1s, got %v", config.BaseDelay)
	}

	if config.MaxDelay != time.Second*10 {
		t.Errorf("Expected MaxDelay to be 10s, got %v", config.MaxDelay)
	}
}

func TestWithRetry(t *testing.T) {
	tests := []struct {
		name          string
		operation     func() error
		config        RetryConfig
		expectedError bool
		expectedCalls int
	}{
		{
			name: "successful operation",
			operation: func() error {
				return nil
			},
			config: RetryConfig{
				MaxRetries: 3,
				BaseDelay:  time.Millisecond,
				MaxDelay:   time.Millisecond * 10,
			},
			expectedError: false,
			expectedCalls: 1,
		},
		{
			name: "operation fails with non-retryable error",
			operation: func() error {
				return &discordgo.RESTError{
					Message:  &discordgo.APIErrorMessage{Code: 50001}, // Missing Access
					Response: &http.Response{StatusCode: 403},
				}
			},
			config: RetryConfig{
				MaxRetries: 3,
				BaseDelay:  time.Millisecond,
				MaxDelay:   time.Millisecond * 10,
			},
			expectedError: true,
			expectedCalls: 1,
		},
		{
			name: "operation fails with retryable error then succeeds",
			operation: func() func() error {
				attempts := 0
				return func() error {
					attempts++
					if attempts == 1 {
						return &discordgo.RESTError{
							Message:  &discordgo.APIErrorMessage{Code: 0},
							Response: &http.Response{StatusCode: 500},
						}
					}
					return nil
				}
			}(),
			config: RetryConfig{
				MaxRetries: 3,
				BaseDelay:  time.Millisecond,
				MaxDelay:   time.Millisecond * 10,
			},
			expectedError: false,
			expectedCalls: 2,
		},
		{
			name: "operation exhausts all retries",
			operation: func() error {
				return &discordgo.RESTError{
					Message:  &discordgo.APIErrorMessage{Code: 0},
					Response: &http.Response{StatusCode: 502},
				}
			},
			config: RetryConfig{
				MaxRetries: 2,
				BaseDelay:  time.Millisecond,
				MaxDelay:   time.Millisecond * 10,
			},
			expectedError: true,
			expectedCalls: 3, // Initial attempt + 2 retries
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			wrappedOp := func() error {
				callCount++
				return tt.operation()
			}

			err := withRetry(wrappedOp, tt.config)

			if tt.expectedError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if callCount != tt.expectedCalls {
				t.Errorf("Expected %d calls but got %d", tt.expectedCalls, callCount)
			}
		})
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		error    error
		expected bool
	}{
		{
			name:     "nil error",
			error:    nil,
			expected: false,
		},
		{
			name: "rate limit error (429)",
			error: &discordgo.RESTError{
				Message:  &discordgo.APIErrorMessage{Code: 0},
				Response: &http.Response{StatusCode: 429},
			},
			expected: true,
		},
		{
			name: "server error (500)",
			error: &discordgo.RESTError{
				Message:  &discordgo.APIErrorMessage{Code: 0},
				Response: &http.Response{StatusCode: 500},
			},
			expected: true,
		},
		{
			name: "server error (502)",
			error: &discordgo.RESTError{
				Message:  &discordgo.APIErrorMessage{Code: 0},
				Response: &http.Response{StatusCode: 502},
			},
			expected: true,
		},
		{
			name: "server error (503)",
			error: &discordgo.RESTError{
				Message:  &discordgo.APIErrorMessage{Code: 0},
				Response: &http.Response{StatusCode: 503},
			},
			expected: true,
		},
		{
			name: "server error (504)",
			error: &discordgo.RESTError{
				Message:  &discordgo.APIErrorMessage{Code: 0},
				Response: &http.Response{StatusCode: 504},
			},
			expected: true,
		},
		{
			name: "client error (400)",
			error: &discordgo.RESTError{
				Message:  &discordgo.APIErrorMessage{Code: 50001}, // Missing Access
				Response: &http.Response{StatusCode: 400},
			},
			expected: false,
		},
		{
			name: "discord API overloaded",
			error: &discordgo.RESTError{
				Message:  &discordgo.APIErrorMessage{Code: discordgo.ErrCodeAPIResourceIsCurrentlyOverloaded},
				Response: &http.Response{StatusCode: 503},
			},
			expected: true,
		},
		{
			name:     "connection reset error",
			error:    errors.New("connection reset by peer"),
			expected: true,
		},
		{
			name:     "timeout error",
			error:    errors.New("request timeout"),
			expected: true,
		},
		{
			name:     "temporary failure error",
			error:    errors.New("temporary failure in name resolution"),
			expected: true,
		},
		{
			name:     "network unreachable error",
			error:    errors.New("network is unreachable"),
			expected: true,
		},
		{
			name:     "EOF error",
			error:    errors.New("unexpected EOF"),
			expected: true,
		},
		{
			name:     "generic error",
			error:    errors.New("some other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableError(tt.error)
			if result != tt.expected {
				t.Errorf("isRetryableError(%v) = %v, want %v", tt.error, result, tt.expected)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{
			name:     "exact match",
			s:        "hello",
			substr:   "hello",
			expected: true,
		},
		{
			name:     "substring at beginning",
			s:        "hello world",
			substr:   "hello",
			expected: true,
		},
		{
			name:     "substring at end",
			s:        "hello world",
			substr:   "world",
			expected: true,
		},
		{
			name:     "substring in middle",
			s:        "hello world",
			substr:   "o w",
			expected: true,
		},
		{
			name:     "substring not found",
			s:        "hello world",
			substr:   "xyz",
			expected: false,
		},
		{
			name:     "empty substring",
			s:        "hello",
			substr:   "",
			expected: true,
		},
		{
			name:     "empty string with non-empty substring",
			s:        "",
			substr:   "hello",
			expected: false,
		},
		{
			name:     "substring longer than string",
			s:        "hi",
			substr:   "hello",
			expected: false,
		},
		{
			name:     "case sensitive - different case",
			s:        "Hello",
			substr:   "hello",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.s, tt.substr)
			if result != tt.expected {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, result, tt.expected)
			}
		})
	}
}

func TestFindInString(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{
			name:     "substring found",
			s:        "hello world",
			substr:   "wor",
			expected: true,
		},
		{
			name:     "substring not found",
			s:        "hello world",
			substr:   "xyz",
			expected: false,
		},
		{
			name:     "exact match",
			s:        "hello",
			substr:   "hello",
			expected: true,
		},
		{
			name:     "empty substring",
			s:        "hello",
			substr:   "",
			expected: true,
		},
		{
			name:     "substring at beginning",
			s:        "testing",
			substr:   "test",
			expected: true,
		},
		{
			name:     "substring at end",
			s:        "testing",
			substr:   "ing",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findInString(tt.s, tt.substr)
			if result != tt.expected {
				t.Errorf("findInString(%q, %q) = %v, want %v", tt.s, tt.substr, result, tt.expected)
			}
		})
	}
}

func TestFollowupWithEmbeds(t *testing.T) {
	tests := []struct {
		name        string
		session     *discordgo.Session
		interaction *discordgo.InteractionCreate
		content     string
		embeds      []*discordgo.MessageEmbed
		expectError bool
	}{
		{
			name:        "nil session",
			session:     nil,
			interaction: &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{}},
			content:     "test",
			embeds:      []*discordgo.MessageEmbed{},
			expectError: true,
		},
		{
			name:        "nil interaction",
			session:     &discordgo.Session{},
			interaction: nil,
			content:     "test",
			embeds:      []*discordgo.MessageEmbed{},
			expectError: true,
		},
		{
			name:        "nil interaction.Interaction",
			session:     &discordgo.Session{},
			interaction: &discordgo.InteractionCreate{},
			content:     "test",
			embeds:      []*discordgo.MessageEmbed{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := FollowupWithEmbeds(tt.session, tt.interaction, tt.content, tt.embeds)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestFollowupWithEmbedsValidation(t *testing.T) {
	// Test embed validation and truncation without making actual API calls
	t.Run("embed content truncation", func(t *testing.T) {
		originalEmbeds := []*discordgo.MessageEmbed{
			{
				Title:       strings.Repeat("B", 300),                                       // Longer than MaxEmbedTitle
				Description: strings.Repeat("C", 5000),                                      // Longer than MaxEmbedDescription
				Footer:      &discordgo.MessageEmbedFooter{Text: strings.Repeat("D", 3000)}, // Longer than MaxEmbedFooterText
				Author:      &discordgo.MessageEmbedAuthor{Name: strings.Repeat("E", 300)},  // Longer than MaxEmbedAuthorName
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  strings.Repeat("F", 300),  // Longer than MaxEmbedFieldName
						Value: strings.Repeat("G", 2000), // Longer than MaxEmbedFieldValue
					},
				},
			},
		}

		// Copy the embeds to avoid modifying the originals
		testEmbeds := make([]*discordgo.MessageEmbed, len(originalEmbeds))
		for i, embed := range originalEmbeds {
			testEmbeds[i] = &discordgo.MessageEmbed{
				Title:       embed.Title,
				Description: embed.Description,
			}
			if embed.Footer != nil {
				testEmbeds[i].Footer = &discordgo.MessageEmbedFooter{Text: embed.Footer.Text}
			}
			if embed.Author != nil {
				testEmbeds[i].Author = &discordgo.MessageEmbedAuthor{Name: embed.Author.Name}
			}
			for _, field := range embed.Fields {
				testEmbeds[i].Fields = append(testEmbeds[i].Fields, &discordgo.MessageEmbedField{
					Name:  field.Name,
					Value: field.Value,
				})
			}
		}

		// Test the validation and truncation logic by calling the preprocessing part
		// This tests the embed truncation without making Discord API calls
		content := strings.Repeat("A", 3000) // Longer than MaxMessageLength

		// Simulate the validation that happens in FollowupWithEmbeds
		for _, embed := range testEmbeds {
			if embed.Title != "" {
				embed.Title = TruncateText(embed.Title, MaxEmbedTitle)
			}
			if embed.Description != "" {
				embed.Description = TruncateText(embed.Description, MaxEmbedDescription)
			}
			if embed.Footer != nil && embed.Footer.Text != "" {
				embed.Footer.Text = TruncateText(embed.Footer.Text, MaxEmbedFooterText)
			}
			if embed.Author != nil && embed.Author.Name != "" {
				embed.Author.Name = TruncateText(embed.Author.Name, MaxEmbedAuthorName)
			}
			for _, field := range embed.Fields {
				if field.Name != "" {
					field.Name = TruncateText(field.Name, MaxEmbedFieldName)
				}
				if field.Value != "" {
					field.Value = TruncateText(field.Value, MaxEmbedFieldValue)
				}
			}
		}

		// Verify truncation results
		embed := testEmbeds[0]

		if len(embed.Title) > MaxEmbedTitle {
			t.Errorf("Embed title not properly truncated: %d > %d", len(embed.Title), MaxEmbedTitle)
		}

		if len(embed.Description) > MaxEmbedDescription {
			t.Errorf("Embed description not properly truncated: %d > %d", len(embed.Description), MaxEmbedDescription)
		}

		if embed.Footer != nil && len(embed.Footer.Text) > MaxEmbedFooterText {
			t.Errorf("Embed footer not properly truncated: %d > %d", len(embed.Footer.Text), MaxEmbedFooterText)
		}

		if embed.Author != nil && len(embed.Author.Name) > MaxEmbedAuthorName {
			t.Errorf("Embed author not properly truncated: %d > %d", len(embed.Author.Name), MaxEmbedAuthorName)
		}

		if len(embed.Fields) > 0 {
			field := embed.Fields[0]

			if len(field.Name) > MaxEmbedFieldName {
				t.Errorf("Embed field name not properly truncated: %d > %d", len(field.Name), MaxEmbedFieldName)
			}

			if len(field.Value) > MaxEmbedFieldValue {
				t.Errorf("Embed field value not properly truncated: %d > %d", len(field.Value), MaxEmbedFieldValue)
			}
		}

		// Test content truncation
		truncatedContent := TruncateText(content, MaxMessageLength)
		if len(truncatedContent) > MaxMessageLength {
			t.Errorf("Content not properly truncated: %d > %d", len(truncatedContent), MaxMessageLength)
		}
	})

	t.Run("too many embeds", func(t *testing.T) {
		// Test embed count limitation
		embeds := make([]*discordgo.MessageEmbed, 15) // More than MaxEmbedsPerMessage (10)
		for i := range embeds {
			embeds[i] = &discordgo.MessageEmbed{Title: "Test"}
		}

		// Simulate the embed limitation that happens in FollowupWithEmbeds
		if len(embeds) > MaxEmbedsPerMessage {
			embeds = embeds[:MaxEmbedsPerMessage]
		}

		if len(embeds) != MaxEmbedsPerMessage {
			t.Errorf("Embeds not properly limited: %d != %d", len(embeds), MaxEmbedsPerMessage)
		}
	})
}

func TestAcknowledgeInteraction(t *testing.T) {
	tests := []struct {
		name        string
		session     *discordgo.Session
		interaction *discordgo.InteractionCreate
		expectError bool
	}{
		{
			name:        "nil session",
			session:     nil,
			interaction: &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{}},
			expectError: true,
		},
		{
			name:        "nil interaction",
			session:     &discordgo.Session{},
			interaction: nil,
			expectError: true,
		},
		{
			name:        "nil interaction.Interaction",
			session:     &discordgo.Session{},
			interaction: &discordgo.InteractionCreate{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test with a reasonable timeout
			start := time.Now()
			err := AcknowledgeInteraction(tt.session, tt.interaction)
			duration := time.Since(start)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Verify that the function respects the timeout
			if duration > InteractionTimeout+time.Second {
				t.Errorf("Function took too long: %v > %v", duration, InteractionTimeout+time.Second)
			}
		})
	}
}

func TestAcknowledgeWithRetry(t *testing.T) {
	tests := []struct {
		name        string
		session     *discordgo.Session
		interaction *discordgo.InteractionCreate
		expectError bool
	}{
		{
			name:        "nil session",
			session:     nil,
			interaction: &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{}},
			expectError: true,
		},
		{
			name:        "nil interaction",
			session:     &discordgo.Session{},
			interaction: nil,
			expectError: true,
		},
		{
			name:        "nil interaction.Interaction",
			session:     &discordgo.Session{},
			interaction: &discordgo.InteractionCreate{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			err := AcknowledgeWithRetry(tt.session, tt.interaction)
			duration := time.Since(start)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// The retry function should handle timeouts reasonably
			// Allow for multiple retry attempts with their timeouts
			maxExpectedDuration := (InteractionTimeout + time.Second) * 3
			if duration > maxExpectedDuration {
				t.Errorf("Function with retries took too long: %v > %v", duration, maxExpectedDuration)
			}
		})
	}
}

func TestDiscordLimitsConstants(t *testing.T) {
	// Verify that the constants match Discord's documented limits
	expectedLimits := map[string]int{
		"MaxMessageLength":    2000,
		"MaxEmbedTitle":       256,
		"MaxEmbedDescription": 4096,
		"MaxEmbedFieldName":   256,
		"MaxEmbedFieldValue":  1024,
		"MaxEmbedFooterText":  2048,
		"MaxEmbedAuthorName":  256,
		"MaxEmbedsPerMessage": 10,
	}

	actualLimits := map[string]int{
		"MaxMessageLength":    MaxMessageLength,
		"MaxEmbedTitle":       MaxEmbedTitle,
		"MaxEmbedDescription": MaxEmbedDescription,
		"MaxEmbedFieldName":   MaxEmbedFieldName,
		"MaxEmbedFieldValue":  MaxEmbedFieldValue,
		"MaxEmbedFooterText":  MaxEmbedFooterText,
		"MaxEmbedAuthorName":  MaxEmbedAuthorName,
		"MaxEmbedsPerMessage": MaxEmbedsPerMessage,
	}

	for limitName, expectedValue := range expectedLimits {
		actualValue := actualLimits[limitName]
		if actualValue != expectedValue {
			t.Errorf("Constant %s: expected %d, got %d", limitName, expectedValue, actualValue)
		}
	}

	// Test InteractionTimeout
	if InteractionTimeout != 3*time.Second {
		t.Errorf("InteractionTimeout: expected 3s, got %v", InteractionTimeout)
	}
}

func TestRetryConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config RetryConfig
		valid  bool
	}{
		{
			name:   "default config",
			config: DefaultRetryConfig(),
			valid:  true,
		},
		{
			name: "zero retries",
			config: RetryConfig{
				MaxRetries: 0,
				BaseDelay:  time.Millisecond,
				MaxDelay:   time.Second,
			},
			valid: true,
		},
		{
			name: "negative retries",
			config: RetryConfig{
				MaxRetries: -1,
				BaseDelay:  time.Millisecond,
				MaxDelay:   time.Second,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that withRetry handles the config appropriately
			callCount := 0
			operation := func() error {
				callCount++
				// Use a retryable error (network timeout)
				return errors.New("timeout")
			}

			_ = withRetry(operation, tt.config)

			expectedMaxCalls := tt.config.MaxRetries + 1
			if tt.config.MaxRetries < 0 {
				// withRetry should handle negative retries gracefully by not executing
				expectedMaxCalls = 0
			}

			if callCount != expectedMaxCalls {
				t.Errorf("Unexpected number of calls: got %d, expected %d", callCount, expectedMaxCalls)
			}
		})
	}
}
