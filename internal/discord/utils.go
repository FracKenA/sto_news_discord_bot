// Package discord provides utility functions and constants for Discord integration in STOBot.
//
// It includes text truncation, response helpers, and Discord API limits.
package discord

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

// Discord API limits constants
const (
	MaxMessageLength    = 2000
	MaxEmbedTitle       = 256
	MaxEmbedDescription = 4096
	MaxEmbedFieldName   = 256
	MaxEmbedFieldValue  = 1024
	MaxEmbedFooterText  = 2048
	MaxEmbedAuthorName  = 256
	MaxEmbedsPerMessage = 10
	InteractionTimeout  = 3 * time.Second // Discord's 3-second acknowledgment requirement
)

// RetryConfig defines retry behavior for Discord API calls
type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

// DefaultRetryConfig returns the default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		BaseDelay:  time.Second,
		MaxDelay:   time.Second * 10,
	}
}

// withRetry executes a function with exponential backoff retry logic
func withRetry(operation func() error, config RetryConfig) error {
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(attempt) * config.BaseDelay
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
			log.Warnf("Retrying Discord operation in %v (attempt %d/%d)", delay, attempt, config.MaxRetries)
			time.Sleep(delay)
		}

		WaitForRateLimit() // Apply rate limiting before each attempt

		if err := operation(); err != nil {
			lastErr = err

			// Check if error is retryable
			if !isRetryableError(err) {
				log.Errorf("Non-retryable Discord error: %v", err)
				return err
			}

			log.Warnf("Retryable Discord error on attempt %d: %v", attempt+1, err)
			continue
		}

		return nil // Success
	}

	log.Errorf("Discord operation failed after %d attempts: %v", config.MaxRetries+1, lastErr)
	return fmt.Errorf("operation failed after %d retries: %w", config.MaxRetries, lastErr)
}

// isRetryableError determines if a Discord API error should be retried
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific Discord error types that are retryable
	if restErr, ok := err.(*discordgo.RESTError); ok {
		// Rate limiting (429) - should always be retried
		if restErr.Response.StatusCode == 429 {
			return true
		}

		// Server errors (5xx) - usually retryable
		switch restErr.Response.StatusCode {
		case 500, 502, 503, 504:
			return true
		}

		// Check specific Discord error codes that are retryable
		switch restErr.Message.Code {
		case discordgo.ErrCodeAPIResourceIsCurrentlyOverloaded:
			return true
		default:
			return false
		}
	}

	// Retry on network-related errors
	errorStr := err.Error()
	retryablePatterns := []string{
		"connection reset",
		"timeout",
		"temporary failure",
		"network is unreachable",
		"EOF",
	}

	for _, pattern := range retryablePatterns {
		if contains(errorStr, pattern) {
			return true
		}
	}

	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					findInString(s, substr)))
}

// findInString performs a simple substring search
func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Respond sends a response to a Discord interaction with retry logic
func Respond(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	if s == nil || i == nil || i.Interaction == nil {
		log.Warn("Cannot respond: nil session or interaction")
		return
	}

	// Truncate content to Discord limits
	content = TruncateText(content, MaxMessageLength)

	operation := func() error {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: content,
				Flags:   discordgo.MessageFlagsEphemeral, // Make command responses private
			},
		})
	}

	if err := withRetry(operation, DefaultRetryConfig()); err != nil {
		log.Errorf("Failed to respond to interaction after retries: %v", err)
	}
}

// RespondError sends an error response to a Discord interaction
func RespondError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	Respond(s, i, fmt.Sprintf("❌ Error: %s", message))
}

// Followup sends a follow-up message to a Discord interaction with retry logic
func Followup(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	if s == nil || i == nil || i.Interaction == nil {
		log.Warn("Cannot send followup: nil session or interaction")
		return
	}

	// Truncate content to Discord limits
	content = TruncateText(content, MaxMessageLength)

	operation := func() error {
		_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral, // Make followup messages private
		})
		return err
	}

	if err := withRetry(operation, DefaultRetryConfig()); err != nil {
		log.Errorf("Failed to send followup message after retries: %v", err)
	}
}

// FollowupError sends an error follow-up message to a Discord interaction
func FollowupError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	Followup(s, i, fmt.Sprintf("❌ Error: %s", message))
}

// FollowupWithEmbeds sends a follow-up message with embeds and retry logic
func FollowupWithEmbeds(s *discordgo.Session, i *discordgo.InteractionCreate, content string, embeds []*discordgo.MessageEmbed) error {
	if s == nil || i == nil || i.Interaction == nil {
		log.Warn("Cannot send followup with embeds: nil session or interaction")
		return fmt.Errorf("nil session or interaction")
	}

	// Validate and truncate embed content
	for _, embed := range embeds {
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

	// Limit number of embeds per message
	if len(embeds) > MaxEmbedsPerMessage {
		embeds = embeds[:MaxEmbedsPerMessage]
		log.Warnf("Truncated embeds to Discord limit of %d", MaxEmbedsPerMessage)
	}

	// Truncate content to Discord limits
	if content != "" {
		content = TruncateText(content, MaxMessageLength)
	}

	operation := func() error {
		_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: content,
			Embeds:  embeds,
			Flags:   discordgo.MessageFlagsEphemeral, // Make followup embeds private
		})
		return err
	}

	return withRetry(operation, DefaultRetryConfig())
}

// TruncateText truncates text to a maximum length, adding ellipsis if needed
func TruncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}

	if maxLength <= 3 {
		// Return truncated ellipsis to fit within maxLength
		return strings.Repeat(".", maxLength)
	}

	return text[:maxLength-3] + "..."
}

// AcknowledgeInteraction safely acknowledges an interaction within Discord's 3-second limit
func AcknowledgeInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if s == nil || i == nil || i.Interaction == nil {
		return fmt.Errorf("nil session or interaction")
	}

	// Use context with timeout to ensure we respect Discord's limits
	ctx, cancel := context.WithTimeout(context.Background(), InteractionTimeout)
	defer cancel()

	// Channel to receive the operation result
	resultChan := make(chan error, 1)

	// Apply rate limiting
	WaitForRateLimit()

	// Perform the acknowledgment in a goroutine
	go func() {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral, // Make deferred responses private
			},
		})
		resultChan <- err
	}()

	// Wait for either completion or timeout
	select {
	case err := <-resultChan:
		if err != nil {
			log.Errorf("Failed to acknowledge interaction: %v", err)
			return err
		}
		log.Debug("Interaction acknowledged successfully")
		return nil
	case <-ctx.Done():
		log.Error("Interaction acknowledgment timed out")
		return fmt.Errorf("interaction acknowledgment timed out")
	}
}

// AcknowledgeWithRetry acknowledges an interaction with retry logic for better reliability
func AcknowledgeWithRetry(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	operation := func() error {
		return AcknowledgeInteraction(s, i)
	}

	config := RetryConfig{
		MaxRetries: 2, // Limited retries for acknowledgment due to time constraints
		BaseDelay:  time.Millisecond * 100,
		MaxDelay:   time.Millisecond * 500,
	}

	return withRetry(operation, config)
}
