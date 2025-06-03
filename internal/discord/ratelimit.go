// Package discord provides rate limiting utilities for Discord API compliance.
//
// This package implements proper rate limiting to ensure the bot follows Discord's
// API rate limits and best practices for API usage. It includes both global rate
// limiting and per-route rate limiting as recommended by Discord.
package discord

import (
	"context"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// Discord rate limit constants based on Discord API documentation
const (
	GlobalRateLimit      = 50   // Global requests per second
	InteractionRateLimit = 3000 // Interactions per second per bot
	MessageRateLimit     = 5    // Messages per 5 seconds per channel
	BulkDeleteRateLimit  = 1    // Bulk deletes per second
	GatewayRateLimit     = 120  // Gateway connects per minute
)

// RateLimiter manages rate limiting for Discord API requests with Discord-specific limits
type RateLimiter struct {
	mu             sync.RWMutex
	lastRequest    time.Time
	minInterval    time.Duration
	globalLimiter  chan struct{}
	requestCount   int
	windowStart    time.Time
	maxRequests    int
	windowDuration time.Duration
}

// RateLimitConfig defines configuration for rate limiting
type RateLimitConfig struct {
	MaxRequests    int           // Maximum requests per window
	WindowDuration time.Duration // Time window duration
	MinInterval    time.Duration // Minimum interval between requests
	MaxConcurrent  int           // Maximum concurrent requests
}

// DefaultRateLimitConfig returns Discord-appropriate rate limiting configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		MaxRequests:    GlobalRateLimit,       // 50 requests per second (Discord global limit)
		WindowDuration: time.Second,           // 1 second window
		MinInterval:    time.Millisecond * 20, // 20ms minimum between requests
		MaxConcurrent:  10,                    // Max 10 concurrent requests
	}
}

// InteractionRateLimitConfig returns configuration optimized for Discord interactions
func InteractionRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		MaxRequests:    20,                    // Conservative limit for interactions
		WindowDuration: time.Second,           // 1 second window
		MinInterval:    time.Millisecond * 50, // 50ms minimum between interactions
		MaxConcurrent:  5,                     // Max 5 concurrent interactions
	}
}

// NewRateLimiter creates a new rate limiter with Discord-appropriate settings
func NewRateLimiter() *RateLimiter {
	config := DefaultRateLimitConfig()
	return NewRateLimiterWithConfig(config)
}

// NewRateLimiterWithConfig creates a rate limiter with custom configuration
func NewRateLimiterWithConfig(config RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		minInterval:    config.MinInterval,
		globalLimiter:  make(chan struct{}, config.MaxConcurrent),
		maxRequests:    config.MaxRequests,
		windowDuration: config.WindowDuration,
		windowStart:    time.Now(),
	}
}

// Wait blocks until it's safe to make another Discord API request
func (rl *RateLimiter) Wait() error {
	return rl.WaitWithContext(context.Background())
}

// WaitWithContext blocks until it's safe to make another Discord API request with context support
func (rl *RateLimiter) WaitWithContext(ctx context.Context) error {
	// Acquire global limiter token with context
	select {
	case rl.globalLimiter <- struct{}{}:
		defer func() { <-rl.globalLimiter }()
	case <-ctx.Done():
		return ctx.Err()
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Reset window if expired
	if now.Sub(rl.windowStart) >= rl.windowDuration {
		rl.windowStart = now
		rl.requestCount = 0
	}

	// Check if we've exceeded the rate limit for this window
	if rl.requestCount >= rl.maxRequests {
		waitTime := rl.windowDuration - now.Sub(rl.windowStart)
		if waitTime > 0 {
			log.Debugf("Rate limit reached, waiting %v before next request", waitTime)

			// Wait with context support
			timer := time.NewTimer(waitTime)
			defer timer.Stop()

			select {
			case <-timer.C:
				// Reset for new window
				rl.windowStart = time.Now()
				rl.requestCount = 0
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	// Check minimum interval between requests
	elapsed := now.Sub(rl.lastRequest)
	if elapsed < rl.minInterval {
		waitTime := rl.minInterval - elapsed
		log.Debugf("Minimum interval not met, waiting %v before next request", waitTime)

		timer := time.NewTimer(waitTime)
		defer timer.Stop()

		select {
		case <-timer.C:
			// Continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Update counters
	rl.lastRequest = time.Now()
	rl.requestCount++

	return nil
}

// GetStats returns current rate limiter statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return map[string]interface{}{
		"requests_in_window": rl.requestCount,
		"max_requests":       rl.maxRequests,
		"window_start":       rl.windowStart,
		"window_duration":    rl.windowDuration,
		"last_request":       rl.lastRequest,
		"min_interval":       rl.minInterval,
	}
}

// Global rate limiter instances for different use cases
var (
	globalRateLimiter      = NewRateLimiter()
	interactionRateLimiter = NewRateLimiterWithConfig(InteractionRateLimitConfig())
)

// WaitForRateLimit waits for the global rate limiter before making Discord API calls
func WaitForRateLimit() {
	if err := globalRateLimiter.Wait(); err != nil {
		log.Errorf("Rate limit wait interrupted: %v", err)
	}
}

// WaitForInteractionRateLimit waits for the interaction-specific rate limiter
func WaitForInteractionRateLimit() {
	if err := interactionRateLimiter.Wait(); err != nil {
		log.Errorf("Interaction rate limit wait interrupted: %v", err)
	}
}

// WaitForRateLimitWithContext waits for rate limit with context support
func WaitForRateLimitWithContext(ctx context.Context) error {
	return globalRateLimiter.WaitWithContext(ctx)
}

// GetGlobalRateLimitStats returns statistics for the global rate limiter
func GetGlobalRateLimitStats() map[string]interface{} {
	return globalRateLimiter.GetStats()
}

// GetInteractionRateLimitStats returns statistics for the interaction rate limiter
func GetInteractionRateLimitStats() map[string]interface{} {
	return interactionRateLimiter.GetStats()
}
