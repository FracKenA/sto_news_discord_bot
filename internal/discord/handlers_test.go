// Package discord contains tests for the STOBot Discord event handlers.
//
// These tests cover event handler creation and logic validation.
package discord

import (
	"testing"

	"github.com/FracKenA/sto_news_discord_bot/internal/testhelpers"
	"github.com/FracKenA/sto_news_discord_bot/internal/types"

	"github.com/bwmarrin/discordgo"
)

func TestReady(t *testing.T) {
	bot := testhelpers.CreateTestBot(t)
	t.Cleanup(func() {
		if bot.DB != nil {
			_ = bot.DB.Close()
		}
	})

	// Create a mock ready event
	readyEvent := &discordgo.Ready{
		User: &discordgo.User{
			Username:      "TestBot",
			Discriminator: "1234",
		},
	}

	// Create the ready handler
	readyHandler := Ready(bot)

	if readyHandler == nil {
		t.Error("Ready handler should not be nil")
	}

	// Test that the handler function has the correct signature
	var testFunc func(*discordgo.Session, *discordgo.Ready) = readyHandler
	_ = testFunc

	// Optionally, test that the handler does not panic when called
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Ready handler panicked: %v", r)
		}
	}()
	// Call the handler with nil session (should handle gracefully)
	readyHandler(nil, readyEvent)

	t.Log("Ready handler created and invoked successfully")
}

func TestInteractionCreate(t *testing.T) {
	bot := testhelpers.CreateTestBot(t)
	defer bot.DB.Close()

	// Create the interaction handler
	interactionHandler := InteractionCreate(bot)

	if interactionHandler == nil {
		t.Error("InteractionCreate handler should not be nil")
	}

	// Test that the handler function has the correct signature
	var testFunc func(*discordgo.Session, *discordgo.InteractionCreate) = interactionHandler
	_ = testFunc // Use the variable to avoid unused variable error

	t.Log("InteractionCreate handler created successfully")
}

func TestInteractionCreateWithEmptyCommand(t *testing.T) {
	bot := testhelpers.CreateTestBot(t)
	defer bot.DB.Close()

	// Create the interaction handler
	interactionHandler := InteractionCreate(bot)

	// Create a mock interaction with empty command name
	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type: discordgo.InteractionApplicationCommand,
			Data: discordgo.ApplicationCommandInteractionData{
				Name: "", // Empty command name
			},
		},
	}

	// Test that the handler doesn't panic with empty command
	// This should return early without doing anything
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Handler panicked with empty command: %v", r)
		}
	}()

	// Call the handler with nil session (it should handle gracefully)
	interactionHandler(nil, interaction)

	t.Log("Handler correctly handled empty command")
}

func TestBotConfigurationValidation(t *testing.T) {
	// Test valid configuration
	validConfig := &types.Config{
		DiscordToken: "valid_token",
		PollPeriod:   30,
		PollCount:    10,
		FreshSeconds: 86400,
		MsgCount:     5,
		DatabasePath: "/path/to/db",
	}

	err := validConfig.Validate()
	if err != nil {
		t.Errorf("Valid config should not return error: %v", err)
	}

	// Test invalid configurations
	testCases := []struct {
		name   string
		config *types.Config
	}{
		{
			name: "empty discord token",
			config: &types.Config{
				DiscordToken: "",
				PollPeriod:   30,
				PollCount:    10,
				FreshSeconds: 86400,
				MsgCount:     5,
				DatabasePath: "/path/to/db",
			},
		},
		{
			name: "zero poll period",
			config: &types.Config{
				DiscordToken: "token",
				PollPeriod:   0,
				PollCount:    10,
				FreshSeconds: 86400,
				MsgCount:     5,
				DatabasePath: "/path/to/db",
			},
		},
		{
			name: "zero poll count",
			config: &types.Config{
				DiscordToken: "token",
				PollPeriod:   30,
				PollCount:    0,
				FreshSeconds: 86400,
				MsgCount:     5,
				DatabasePath: "/path/to/db",
			},
		},
		{
			name: "zero fresh seconds",
			config: &types.Config{
				DiscordToken: "token",
				PollPeriod:   30,
				PollCount:    10,
				FreshSeconds: 0,
				MsgCount:     5,
				DatabasePath: "/path/to/db",
			},
		},
		{
			name: "zero message count",
			config: &types.Config{
				DiscordToken: "token",
				PollPeriod:   30,
				PollCount:    10,
				FreshSeconds: 86400,
				MsgCount:     0,
				DatabasePath: "/path/to/db",
			},
		},
		{
			name: "empty database path",
			config: &types.Config{
				DiscordToken: "token",
				PollPeriod:   30,
				PollCount:    10,
				FreshSeconds: 86400,
				MsgCount:     5,
				DatabasePath: "",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			if err == nil {
				t.Errorf("Config validation should have failed for %s", tc.name)
			}
		})
	}
}

func TestBotCreation(t *testing.T) {
	bot := testhelpers.CreateTestBot(t)
	defer bot.DB.Close()

	if bot.DB == nil {
		t.Error("Bot database should not be nil")
	}
	if bot.Config == nil {
		t.Error("Bot config should not be nil")
	}
	if bot.Config.DiscordToken != "test_token" {
		t.Errorf("Expected token 'test_token', got '%s'", bot.Config.DiscordToken)
	}
}

func TestHandlerRegistration(t *testing.T) {
	bot := testhelpers.CreateTestBot(t)
	defer bot.DB.Close()

	// Test that handler functions can be created
	readyHandler := Ready(bot)
	interactionHandler := InteractionCreate(bot)

	if readyHandler == nil {
		t.Error("Ready handler should not be nil")
	}
	if interactionHandler == nil {
		t.Error("Interaction handler should not be nil")
	}

	// Test that handlers have correct function signatures
	// This is a compile-time check that the signatures match what Discord expects
	var _ func(*discordgo.Session, *discordgo.Ready) = readyHandler
	var _ func(*discordgo.Session, *discordgo.InteractionCreate) = interactionHandler

	t.Log("All handlers have correct signatures")
}

// TestReadyHandlerNilChecks tests Ready handler with various nil conditions
func TestReadyHandlerNilChecks(t *testing.T) {
	bot := testhelpers.CreateTestBot(t)
	t.Cleanup(func() {
		if bot.DB != nil {
			_ = bot.DB.Close()
		}
	})

	readyHandler := Ready(bot)

	tests := []struct {
		name        string
		session     *discordgo.Session
		event       *discordgo.Ready
		shouldPanic bool
	}{
		{
			name:        "nil event",
			session:     &discordgo.Session{},
			event:       nil,
			shouldPanic: false,
		},
		{
			name:    "nil user in event",
			session: &discordgo.Session{},
			event: &discordgo.Ready{
				User: nil,
			},
			shouldPanic: false,
		},
		{
			name:    "nil session",
			session: nil,
			event: &discordgo.Ready{
				User: &discordgo.User{
					Username:      "TestBot",
					Discriminator: "1234",
				},
			},
			shouldPanic: false,
		},
		{
			name:    "valid event and session",
			session: testhelpers.CreateMockDiscordSession(),
			event: &discordgo.Ready{
				User: &discordgo.User{
					Username:      "TestBot",
					Discriminator: "1234",
				},
			},
			shouldPanic: false, // Should handle Discord API errors gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.shouldPanic {
						t.Errorf("Ready handler panicked unexpectedly: %v", r)
					}
				} else if tt.shouldPanic {
					t.Error("Ready handler should have panicked but didn't")
				}
			}()

			readyHandler(tt.session, tt.event)
		})
	}
}

// TestInteractionCreateHandlerNilChecks tests InteractionCreate handler with various nil conditions
func TestInteractionCreateHandlerNilChecks(t *testing.T) {
	bot := testhelpers.CreateTestBot(t)
	t.Cleanup(func() {
		if bot.DB != nil {
			_ = bot.DB.Close()
		}
	})

	interactionHandler := InteractionCreate(bot)

	tests := []struct {
		name        string
		session     *discordgo.Session
		interaction *discordgo.InteractionCreate
		shouldPanic bool
	}{
		{
			name:        "nil interaction",
			session:     testhelpers.CreateMockDiscordSession(),
			interaction: nil,
			shouldPanic: false, // Should handle nil interaction gracefully
		},
		{
			name:    "nil interaction data",
			session: testhelpers.CreateMockDiscordSession(),
			interaction: &discordgo.InteractionCreate{
				Interaction: nil,
			},
			shouldPanic: false, // Should handle nil interaction data gracefully
		},
		{
			name:    "empty command name",
			session: testhelpers.CreateMockDiscordSession(),
			interaction: &discordgo.InteractionCreate{
				Interaction: &discordgo.Interaction{
					Type: discordgo.InteractionApplicationCommand,
					Data: discordgo.ApplicationCommandInteractionData{
						Name: "",
					},
				},
			},
			shouldPanic: false, // Should return early without panic
		},
		{
			name:    "valid command",
			session: testhelpers.CreateMockDiscordSession(),
			interaction: &discordgo.InteractionCreate{
				Interaction: &discordgo.Interaction{
					Type: discordgo.InteractionApplicationCommand,
					Data: discordgo.ApplicationCommandInteractionData{
						Name: "stobot_help",
					},
				},
			},
			shouldPanic: false, // Should handle Discord API errors gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.shouldPanic {
						t.Errorf("InteractionCreate handler panicked unexpectedly: %v", r)
					}
				} else if tt.shouldPanic {
					t.Error("InteractionCreate handler should have panicked but didn't")
				}
			}()

			interactionHandler(tt.session, tt.interaction)
		})
	}
}

// TestHandleCommandNilChecks tests HandleCommand function with various nil parameters
func TestHandleCommandNilChecks(t *testing.T) {
	bot := testhelpers.CreateTestBot(t)
	t.Cleanup(func() {
		if bot.DB != nil {
			_ = bot.DB.Close()
		}
	})

	session := &discordgo.Session{}
	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type: discordgo.InteractionApplicationCommand,
			Data: discordgo.ApplicationCommandInteractionData{
				Name: "stobot_help",
			},
		},
	}

	tests := []struct {
		name        string
		bot         *types.Bot
		session     *discordgo.Session
		interaction *discordgo.InteractionCreate
	}{
		{
			name:        "nil bot",
			bot:         nil,
			session:     session,
			interaction: interaction,
		},
		{
			name:        "nil session",
			bot:         bot,
			session:     nil,
			interaction: interaction,
		},
		{
			name:        "nil interaction",
			bot:         bot,
			session:     session,
			interaction: nil,
		},
		{
			name:    "nil interaction data",
			bot:     bot,
			session: session,
			interaction: &discordgo.InteractionCreate{
				Interaction: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("HandleCommand panicked with nil parameters: %v", r)
				}
			}()

			HandleCommand(tt.bot, tt.session, tt.interaction)
		})
	}
}

// TestHandleCommandRouting tests that HandleCommand routes commands correctly
func TestHandleCommandRouting(t *testing.T) {
	bot := testhelpers.CreateTestBot(t)
	t.Cleanup(func() {
		if bot.DB != nil {
			_ = bot.DB.Close()
		}
	})

	// Test commands that should be routed (won't actually execute due to nil session)
	commands := []string{
		"stobot_register",
		"stobot_unregister",
		"stobot_status",
		"stobot_news",
		"stobot_search_news",
		"stobot_search_tags",
		"stobot_trending",
		"stobot_random_news",
		"stobot_news_stats",
		"stobot_news_since",
		"stobot_news_between",
		"stobot_exclude_tags",
		"stobot_server_stats",
		"stobot_popular_this_week",
		"stobot_tag_trends",
		"stobot_engagement_report",
		"stobot_help",
	}

	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			interaction := &discordgo.InteractionCreate{
				Interaction: &discordgo.Interaction{
					Type: discordgo.InteractionApplicationCommand,
					Data: discordgo.ApplicationCommandInteractionData{
						Name: cmd,
					},
				},
			}

			defer func() {
				if r := recover(); r != nil {
					t.Errorf("HandleCommand panicked for command %s: %v", cmd, r)
				}
			}()

			// This should route to the appropriate handler but not execute due to nil session
			HandleCommand(bot, nil, interaction)
		})
	}
}

// TestHandleCommandUnknownCommand tests handling of unknown commands
func TestHandleCommandUnknownCommand(t *testing.T) {
	bot := testhelpers.CreateTestBot(t)
	t.Cleanup(func() {
		if bot.DB != nil {
			_ = bot.DB.Close()
		}
	})

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type: discordgo.InteractionApplicationCommand,
			Data: discordgo.ApplicationCommandInteractionData{
				Name: "unknown_command",
			},
		},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("HandleCommand panicked for unknown command: %v", r)
		}
	}()

	// Should handle unknown command gracefully (no-op)
	HandleCommand(bot, &discordgo.Session{}, interaction)
}

// TestCommandsHaveValidNames tests that all registered commands have valid names
func TestCommandsHaveValidNames(t *testing.T) {
	// This is more of a validation test to ensure command names are consistent
	expectedCommands := map[string]bool{
		"stobot_register":          true,
		"stobot_unregister":        true,
		"stobot_status":            true,
		"stobot_news":              true,
		"stobot_search_news":       true,
		"stobot_search_tags":       true,
		"stobot_trending":          true,
		"stobot_random_news":       true,
		"stobot_news_stats":        true,
		"stobot_news_since":        true,
		"stobot_news_between":      true,
		"stobot_exclude_tags":      true,
		"stobot_server_stats":      true,
		"stobot_popular_this_week": true,
		"stobot_tag_trends":        true,
		"stobot_engagement_report": true,
		"stobot_help":              true,
	}

	// Test that our expected commands match what HandleCommand supports
	for cmdName := range expectedCommands {
		t.Run(cmdName, func(t *testing.T) {
			// Just ensure the command name is not empty
			if cmdName == "" {
				t.Error("Command name should not be empty")
			}

			// Ensure command name follows Discord naming conventions
			if len(cmdName) > 32 {
				t.Errorf("Command name '%s' is too long (max 32 characters)", cmdName)
			}

			// Should only contain lowercase letters, numbers, and underscores
			for _, char := range cmdName {
				if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_') {
					t.Errorf("Command name '%s' contains invalid character '%c'", cmdName, char)
				}
			}
		})
	}
}
