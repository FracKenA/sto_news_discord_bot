// Package discord contains tests for the STOBot Discord registration handlers.
//
// These tests cover registration, unregistration, and status command handlers.
package discord

import (
	"testing"

	"github.com/FracKenA/sto_news_discord_bot/internal/testhelpers"
	"github.com/FracKenA/sto_news_discord_bot/internal/types"

	"github.com/bwmarrin/discordgo"
)

// TestHandleRegisterNilChecks tests handleRegister with various nil conditions
func TestHandleRegisterNilChecks(t *testing.T) {
	bot := testhelpers.CreateTestBot(t)
	t.Cleanup(func() {
		if bot.DB != nil {
			_ = bot.DB.Close()
		}
	})

	tests := []struct {
		name        string
		bot         *types.Bot
		session     *discordgo.Session
		interaction *discordgo.InteractionCreate
		shouldPanic bool
	}{
		{
			name:        "nil bot",
			bot:         nil,
			session:     testhelpers.CreateMockDiscordSession(),
			interaction: createMockRegisterInteraction(),
			shouldPanic: false,
		},
		{
			name:        "nil session",
			bot:         bot,
			session:     nil,
			interaction: createMockRegisterInteraction(),
			shouldPanic: false,
		},
		{
			name:        "nil interaction",
			bot:         bot,
			session:     testhelpers.CreateMockDiscordSession(),
			interaction: nil,
			shouldPanic: false,
		},
		{
			name:        "valid parameters",
			bot:         bot,
			session:     testhelpers.CreateMockDiscordSession(),
			interaction: createMockRegisterInteraction(),
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.shouldPanic {
						t.Errorf("handleRegister panicked unexpectedly: %v", r)
					}
				} else if tt.shouldPanic {
					t.Error("handleRegister should have panicked but didn't")
				}
			}()

			handleRegister(tt.bot, tt.session, tt.interaction)
		})
	}
}

// TestHandleUnregisterNilChecks tests handleUnregister with various nil conditions
func TestHandleUnregisterNilChecks(t *testing.T) {
	bot := testhelpers.CreateTestBot(t)
	t.Cleanup(func() {
		if bot.DB != nil {
			_ = bot.DB.Close()
		}
	})

	tests := []struct {
		name        string
		bot         *types.Bot
		session     *discordgo.Session
		interaction *discordgo.InteractionCreate
		shouldPanic bool
	}{
		{
			name:        "nil bot",
			bot:         nil,
			session:     testhelpers.CreateMockDiscordSession(),
			interaction: createMockUnregisterInteraction(),
			shouldPanic: false,
		},
		{
			name:        "nil session",
			bot:         bot,
			session:     nil,
			interaction: createMockUnregisterInteraction(),
			shouldPanic: false,
		},
		{
			name:        "nil interaction",
			bot:         bot,
			session:     testhelpers.CreateMockDiscordSession(),
			interaction: nil,
			shouldPanic: false,
		},
		{
			name:        "valid parameters",
			bot:         bot,
			session:     testhelpers.CreateMockDiscordSession(),
			interaction: createMockUnregisterInteraction(),
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.shouldPanic {
						t.Errorf("handleUnregister panicked unexpectedly: %v", r)
					}
				} else if tt.shouldPanic {
					t.Error("handleUnregister should have panicked but didn't")
				}
			}()

			handleUnregister(tt.bot, tt.session, tt.interaction)
		})
	}
}

// TestHandleStatusNilChecks tests handleStatus with various nil conditions
func TestHandleStatusNilChecks(t *testing.T) {
	bot := testhelpers.CreateTestBot(t)
	t.Cleanup(func() {
		if bot.DB != nil {
			_ = bot.DB.Close()
		}
	})

	tests := []struct {
		name        string
		bot         *types.Bot
		session     *discordgo.Session
		interaction *discordgo.InteractionCreate
		shouldPanic bool
	}{
		{
			name:        "nil bot",
			bot:         nil,
			session:     testhelpers.CreateMockDiscordSession(),
			interaction: createMockStatusInteraction(),
			shouldPanic: false,
		},
		{
			name:        "nil session",
			bot:         bot,
			session:     nil,
			interaction: createMockStatusInteraction(),
			shouldPanic: false,
		},
		{
			name:        "nil interaction",
			bot:         bot,
			session:     testhelpers.CreateMockDiscordSession(),
			interaction: nil,
			shouldPanic: false,
		},
		{
			name:        "valid parameters",
			bot:         bot,
			session:     testhelpers.CreateMockDiscordSession(),
			interaction: createMockStatusInteraction(),
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.shouldPanic {
						t.Errorf("handleStatus panicked unexpectedly: %v", r)
					}
				} else if tt.shouldPanic {
					t.Error("handleStatus should have panicked but didn't")
				}
			}()

			handleStatus(tt.bot, tt.session, tt.interaction)
		})
	}
}

// TestHandleRegisterWithOptions tests handleRegister with different platform options
func TestHandleRegisterWithOptions(t *testing.T) {
	bot := testhelpers.CreateTestBot(t)
	t.Cleanup(func() {
		if bot.DB != nil {
			_ = bot.DB.Close()
		}
	})

	tests := []struct {
		name      string
		platforms string
	}{
		{
			name:      "default platforms",
			platforms: "",
		},
		{
			name:      "pc only",
			platforms: "pc",
		},
		{
			name:      "xbox only",
			platforms: "xbox",
		},
		{
			name:      "ps only",
			platforms: "ps",
		},
		{
			name:      "pc and xbox",
			platforms: "pc,xbox",
		},
		{
			name:      "all platforms",
			platforms: "pc,xbox,ps",
		},
		{
			name:      "platforms with spaces",
			platforms: "pc, xbox, ps",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interaction := createMockRegisterInteractionWithPlatforms(tt.platforms)

			defer func() {
				if r := recover(); r != nil {
					t.Errorf("handleRegister panicked with platforms '%s': %v", tt.platforms, r)
				}
			}()

			// This will fail due to nil session, but should not panic
			handleRegister(bot, nil, interaction)
		})
	}
}

// TestRegistrationHandlerPermissions tests permission checking behavior
func TestRegistrationHandlerPermissions(t *testing.T) {
	bot := testhelpers.CreateTestBot(t)
	t.Cleanup(func() {
		if bot.DB != nil {
			_ = bot.DB.Close()
		}
	})

	// Test interactions with different permission scenarios
	tests := []struct {
		name        string
		interaction *discordgo.InteractionCreate
		handler     func(*types.Bot, *discordgo.Session, *discordgo.InteractionCreate)
	}{
		{
			name:        "register without admin permissions",
			interaction: createMockRegisterInteractionNoAdmin(),
			handler:     handleRegister,
		},
		{
			name:        "unregister without admin permissions",
			interaction: createMockUnregisterInteractionNoAdmin(),
			handler:     handleUnregister,
		},
		{
			name:        "status command (no admin required)",
			interaction: createMockStatusInteraction(),
			handler:     handleStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Handler panicked for %s: %v", tt.name, r)
				}
			}()

			// This will fail due to nil session, but should not panic
			tt.handler(bot, nil, tt.interaction)
		})
	}
}

// TestRegistrationHandlerMalformedData tests handlers with malformed interaction data
func TestRegistrationHandlerMalformedData(t *testing.T) {
	bot := testhelpers.CreateTestBot(t)
	t.Cleanup(func() {
		if bot.DB != nil {
			_ = bot.DB.Close()
		}
	})

	tests := []struct {
		name        string
		interaction *discordgo.InteractionCreate
		handler     func(*types.Bot, *discordgo.Session, *discordgo.InteractionCreate)
	}{
		{
			name:        "register with nil interaction data",
			interaction: createMalformedInteraction(),
			handler:     handleRegister,
		},
		{
			name:        "unregister with nil interaction data",
			interaction: createMalformedInteraction(),
			handler:     handleUnregister,
		},
		{
			name:        "status with nil interaction data",
			interaction: createMalformedInteraction(),
			handler:     handleStatus,
		},
		{
			name:        "register with empty channel ID",
			interaction: createEmptyChannelInteraction(),
			handler:     handleRegister,
		},
		{
			name:        "unregister with empty channel ID",
			interaction: createEmptyChannelInteraction(),
			handler:     handleUnregister,
		},
		{
			name:        "status with empty channel ID",
			interaction: createEmptyChannelInteraction(),
			handler:     handleStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Handler panicked for %s: %v", tt.name, r)
				}
			}()

			// These should handle malformed data gracefully
			tt.handler(bot, nil, tt.interaction)
		})
	}
}

// Helper functions to create mock interactions

func createMockRegisterInteraction() *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type:      discordgo.InteractionApplicationCommand,
			ChannelID: "123456789",
			Data: discordgo.ApplicationCommandInteractionData{
				Name:    "stobot_register",
				Options: []*discordgo.ApplicationCommandInteractionDataOption{},
			},
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       "987654321",
					Username: "testuser",
				},
				Permissions: discordgo.PermissionAdministrator,
			},
		},
	}
}

func createMockRegisterInteractionWithPlatforms(platforms string) *discordgo.InteractionCreate {
	interaction := createMockRegisterInteraction()
	if platforms != "" {
		interaction.Interaction.Data = discordgo.ApplicationCommandInteractionData{
			Name: "stobot_register",
			Options: []*discordgo.ApplicationCommandInteractionDataOption{
				{
					Name:  "platforms",
					Type:  discordgo.ApplicationCommandOptionString,
					Value: platforms,
				},
			},
		}
	}
	return interaction
}

func createMockRegisterInteractionNoAdmin() *discordgo.InteractionCreate {
	interaction := createMockRegisterInteraction()
	// Remove admin permissions
	interaction.Interaction.Member.Permissions = 0
	return interaction
}

func createMockUnregisterInteraction() *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type:      discordgo.InteractionApplicationCommand,
			ChannelID: "123456789",
			Data: discordgo.ApplicationCommandInteractionData{
				Name: "stobot_unregister",
			},
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       "987654321",
					Username: "testuser",
				},
				Permissions: discordgo.PermissionAdministrator,
			},
		},
	}
}

func createMockUnregisterInteractionNoAdmin() *discordgo.InteractionCreate {
	interaction := createMockUnregisterInteraction()
	// Remove admin permissions
	interaction.Interaction.Member.Permissions = 0
	return interaction
}

func createMockStatusInteraction() *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type:      discordgo.InteractionApplicationCommand,
			ChannelID: "123456789",
			Data: discordgo.ApplicationCommandInteractionData{
				Name: "stobot_status",
			},
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       "987654321",
					Username: "testuser",
				},
			},
		},
	}
}

func createMalformedInteraction() *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{
		Interaction: nil,
	}
}

func createEmptyChannelInteraction() *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type:      discordgo.InteractionApplicationCommand,
			ChannelID: "", // Empty channel ID
			Data: discordgo.ApplicationCommandInteractionData{
				Name: "stobot_register",
			},
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       "987654321",
					Username: "testuser",
				},
			},
		},
	}
}
