// Package discord provides Discord event handler implementations for STOBot.
//
// It includes Ready, InteractionCreate, and related event logic.
package discord

import (
	"github.com/FracKenA/sto_news_discord_bot/internal/types"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

// Ready handles the ready event when bot connects to Discord
func Ready(b *types.Bot) func(s *discordgo.Session, event *discordgo.Ready) {
	return func(s *discordgo.Session, event *discordgo.Ready) {
		if event == nil || event.User == nil {
			log.Warning("Ready event or user is nil")
			return
		}

		log.Infof("Bot connected as %s#%s", event.User.Username, event.User.Discriminator)

		// Skip Discord API calls if session is nil (for testing)
		if s == nil {
			log.Warning("Session is nil, skipping Discord API calls")
			return
		}

		// Set status
		err := s.UpdateGameStatus(0, "Monitoring Star Trek Online news")
		if err != nil {
			log.Errorf("Failed to set status: %v", err)
		}

		// Register slash commands
		RegisterCommands(s)
		log.Info("Slash commands registered successfully")
	}
}

// InteractionCreate handles slash command interactions
func InteractionCreate(b *types.Bot) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// Check for nil interaction
		if i == nil || i.Interaction == nil {
			log.Warning("Received nil interaction")
			return
		}

		// Check for empty command name
		if i.ApplicationCommandData().Name == "" {
			return
		}

		HandleCommand(b, s, i)
	}
}
