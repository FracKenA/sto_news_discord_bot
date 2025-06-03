// Package discord contains tests for the STOBot Discord news handlers.
//
// These tests cover news command handlers and related functionality.
package discord

import (
	"testing"

	"github.com/FracKenA/sto_news_discord_bot/internal/testhelpers"
	"github.com/FracKenA/sto_news_discord_bot/internal/types"

	"github.com/bwmarrin/discordgo"
)

// TestHandleNewsNilChecks tests handleNews with various nil conditions
func TestHandleNewsNilChecks(t *testing.T) {
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
		tag         string
		shouldPanic bool
	}{
		{
			name:        "nil bot",
			bot:         nil,
			session:     testhelpers.CreateMockDiscordSession(),
			interaction: createMockNewsInteraction(),
			tag:         "star-trek-online",
			shouldPanic: false, // Now handles nil bot gracefully
		},
		{
			name:        "nil session",
			bot:         bot,
			session:     nil,
			interaction: createMockNewsInteraction(),
			tag:         "star-trek-online",
			shouldPanic: false, // Should handle nil session gracefully
		},
		{
			name:        "nil interaction",
			bot:         bot,
			session:     testhelpers.CreateMockDiscordSession(),
			interaction: nil,
			tag:         "star-trek-online",
			shouldPanic: false, // Should handle nil interaction gracefully
		},
		{
			name:        "valid parameters",
			bot:         bot,
			session:     testhelpers.CreateMockDiscordSession(),
			interaction: createMockNewsInteraction(),
			tag:         "star-trek-online",
			shouldPanic: false, // Should handle Discord API errors gracefully
		},
		{
			name:        "empty tag",
			bot:         bot,
			session:     testhelpers.CreateMockDiscordSession(),
			interaction: createMockNewsInteraction(),
			tag:         "",
			shouldPanic: false, // Now handles empty tag gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.shouldPanic {
						t.Errorf("handleNews panicked unexpectedly: %v", r)
					}
				} else if tt.shouldPanic {
					t.Error("handleNews should have panicked but didn't")
				}
			}()

			handleNews(tt.bot, tt.session, tt.interaction, tt.tag)
		})
	}
}

// TestNewsCommandsWithOptions tests news commands with various option combinations
func TestNewsCommandsWithOptions(t *testing.T) {
	bot := testhelpers.CreateTestBot(t)
	t.Cleanup(func() {
		if bot.DB != nil {
			_ = bot.DB.Close()
		}
	})

	// Test news command with different tags
	newsTests := []struct {
		name string
		tag  string
	}{
		{"default tag", "star-trek-online"},
		{"patch notes", "patch-notes"},
		{"events", "events"},
		{"dev blogs", "dev-blogs"},
		{"empty tag", ""},
	}

	for _, tt := range newsTests {
		t.Run("news_"+tt.name, func(t *testing.T) {
			interaction := createMockNewsInteractionWithTag(tt.tag)

			defer func() {
				if r := recover(); r != nil {
					t.Errorf("handleNews panicked with tag '%s': %v", tt.tag, r)
				}
			}()

			handleNews(bot, nil, interaction, tt.tag)
		})
	}
}

// Helper functions to create mock interactions for news commands

func createMockNewsInteraction() *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type:      discordgo.InteractionApplicationCommand,
			ChannelID: "123456789",
			Data: discordgo.ApplicationCommandInteractionData{
				Name: "stobot_news",
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

func createMockNewsInteractionWithTag(tag string) *discordgo.InteractionCreate {
	interaction := createMockNewsInteraction()
	if tag != "" {
		interaction.Interaction.Data = discordgo.ApplicationCommandInteractionData{
			Name: "stobot_news",
			Options: []*discordgo.ApplicationCommandInteractionDataOption{
				{
					Name:  "tag",
					Type:  discordgo.ApplicationCommandOptionString,
					Value: tag,
				},
			},
		}
	}
	return interaction
}
