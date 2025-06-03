// Package discord contains tests for the STOBot Discord statistics handlers.
//
// These tests cover statistics command handlers and related functionality.
package discord

import (
	"testing"

	"github.com/FracKenA/sto_news_discord_bot/internal/testhelpers"
	"github.com/FracKenA/sto_news_discord_bot/internal/types"

	"github.com/bwmarrin/discordgo"
)

// TestHandleNewsStatsNilChecks tests handleNewsStats with various nil conditions
func TestHandleNewsStatsNilChecks(t *testing.T) {
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
			interaction: createMockNewsStatsInteraction(),
			shouldPanic: false,
		},
		{
			name:        "nil session",
			bot:         bot,
			session:     nil,
			interaction: createMockNewsStatsInteraction(),
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
			interaction: createMockNewsStatsInteraction(),
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.shouldPanic {
						t.Errorf("handleNewsStats panicked unexpectedly: %v", r)
					}
				} else if tt.shouldPanic {
					t.Error("handleNewsStats should have panicked but didn't")
				}
			}()

			handleNewsStats(tt.bot, tt.session, tt.interaction)
		})
	}
}

// TestHandleServerStatsNilChecks tests handleServerStats with various nil conditions
func TestHandleServerStatsNilChecks(t *testing.T) {
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
			interaction: createMockServerStatsInteraction(),
			shouldPanic: false,
		},
		{
			name:        "nil session",
			bot:         bot,
			session:     nil,
			interaction: createMockServerStatsInteraction(),
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
			interaction: createMockServerStatsInteraction(),
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.shouldPanic {
						t.Errorf("handleServerStats panicked unexpectedly: %v", r)
					}
				} else if tt.shouldPanic {
					t.Error("handleServerStats should have panicked but didn't")
				}
			}()

			handleServerStats(tt.bot, tt.session, tt.interaction)
		})
	}
}

// TestHandlePopularThisWeekNilChecks tests handlePopularThisWeek with various nil conditions
func TestHandlePopularThisWeekNilChecks(t *testing.T) {
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
			interaction: createMockPopularThisWeekInteraction(),
			shouldPanic: false,
		},
		{
			name:        "nil session",
			bot:         bot,
			session:     nil,
			interaction: createMockPopularThisWeekInteraction(),
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
			interaction: createMockPopularThisWeekInteraction(),
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.shouldPanic {
						t.Errorf("handlePopularThisWeek panicked unexpectedly: %v", r)
					}
				} else if tt.shouldPanic {
					t.Error("handlePopularThisWeek should have panicked but didn't")
				}
			}()

			handlePopularThisWeek(tt.bot, tt.session, tt.interaction)
		})
	}
}

// TestHandleTagTrendsNilChecks tests handleTagTrends with various nil conditions
func TestHandleTagTrendsNilChecks(t *testing.T) {
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
			interaction: createMockTagTrendsInteraction(),
			shouldPanic: false,
		},
		{
			name:        "nil session",
			bot:         bot,
			session:     nil,
			interaction: createMockTagTrendsInteraction(),
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
			interaction: createMockTagTrendsInteraction(),
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.shouldPanic {
						t.Errorf("handleTagTrends panicked unexpectedly: %v", r)
					}
				} else if tt.shouldPanic {
					t.Error("handleTagTrends should have panicked but didn't")
				}
			}()

			handleTagTrends(tt.bot, tt.session, tt.interaction)
		})
	}
}

// TestHandleEngagementReportNilChecks tests handleEngagementReport with various nil conditions
func TestHandleEngagementReportNilChecks(t *testing.T) {
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
			interaction: createMockEngagementReportInteraction(),
			shouldPanic: false,
		},
		{
			name:        "nil session",
			bot:         bot,
			session:     nil,
			interaction: createMockEngagementReportInteraction(),
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
			interaction: createMockEngagementReportInteraction(),
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.shouldPanic {
						t.Errorf("handleEngagementReport panicked unexpectedly: %v", r)
					}
				} else if tt.shouldPanic {
					t.Error("handleEngagementReport should have panicked but didn't")
				}
			}()

			handleEngagementReport(tt.bot, tt.session, tt.interaction)
		})
	}
}

// TestStatsHandlersWithOptions tests stats handlers with various option combinations
func TestStatsHandlersWithOptions(t *testing.T) {
	bot := testhelpers.CreateTestBot(t)
	t.Cleanup(func() {
		if bot.DB != nil {
			_ = bot.DB.Close()
		}
	})

	// Test tag trends with different periods
	tagTrendsTests := []struct {
		name   string
		period string
	}{
		{"default period", ""},
		{"week period", "week"},
		{"month period", "month"},
		{"quarter period", "quarter"},
	}

	for _, tt := range tagTrendsTests {
		t.Run("tag_trends_"+tt.name, func(t *testing.T) {
			interaction := createMockTagTrendsInteractionWithPeriod(tt.period)

			defer func() {
				if r := recover(); r != nil {
					t.Errorf("handleTagTrends panicked with period '%s': %v", tt.period, r)
				}
			}()

			handleTagTrends(bot, nil, interaction)
		})
	}
}

// Helper functions to create mock interactions for stats commands

func createMockNewsStatsInteraction() *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type:      discordgo.InteractionApplicationCommand,
			ChannelID: "123456789",
			Data: discordgo.ApplicationCommandInteractionData{
				Name: "stobot_news_stats",
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

func createMockServerStatsInteraction() *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type:      discordgo.InteractionApplicationCommand,
			ChannelID: "123456789",
			Data: discordgo.ApplicationCommandInteractionData{
				Name: "stobot_server_stats",
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

func createMockPopularThisWeekInteraction() *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type:      discordgo.InteractionApplicationCommand,
			ChannelID: "123456789",
			Data: discordgo.ApplicationCommandInteractionData{
				Name: "stobot_popular_this_week",
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

func createMockTagTrendsInteraction() *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type:      discordgo.InteractionApplicationCommand,
			ChannelID: "123456789",
			Data: discordgo.ApplicationCommandInteractionData{
				Name: "stobot_tag_trends",
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

func createMockTagTrendsInteractionWithPeriod(period string) *discordgo.InteractionCreate {
	interaction := createMockTagTrendsInteraction()
	if period != "" {
		interaction.Interaction.Data = discordgo.ApplicationCommandInteractionData{
			Name: "stobot_tag_trends",
			Options: []*discordgo.ApplicationCommandInteractionDataOption{
				{
					Name:  "period",
					Type:  discordgo.ApplicationCommandOptionString,
					Value: period,
				},
			},
		}
	}
	return interaction
}

func createMockEngagementReportInteraction() *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type:      discordgo.InteractionApplicationCommand,
			ChannelID: "123456789",
			Data: discordgo.ApplicationCommandInteractionData{
				Name: "stobot_engagement_report",
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
