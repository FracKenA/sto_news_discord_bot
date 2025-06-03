package discord

import (
	"fmt"
	"strings"

	"github.com/FracKenA/sto_news_discord_bot/internal/database"
	"github.com/FracKenA/sto_news_discord_bot/internal/types"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

// handleRegister handles the "register" command interaction
func handleRegister(b *types.Bot, s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Validate inputs
	if i == nil || i.Interaction == nil {
		log.Warning("handleRegister called with nil interaction")
		return
	}

	// Check if user has administrator permission
	if !hasAdminPermission(s, i) {
		RespondError(s, i, "You need Administrator permission to use this command.")
		return
	}

	// Acknowledge interaction with timeout handling
	if err := AcknowledgeWithRetry(s, i); err != nil {
		log.Errorf("Failed to acknowledge register command: %v", err)
		return
	}

	data := i.ApplicationCommandData()
	platforms := "pc,xbox,ps" // default

	for _, option := range data.Options {
		if option.Name == "platforms" && option.StringValue() != "" {
			platforms = option.StringValue()
		}
	}

	channelID := i.ChannelID

	err := database.AddChannel(b, channelID)
	if err != nil {
		Followup(s, i, fmt.Sprintf("❌ Failed to register channel: %v", err))
		return
	}

	// Update platforms if specified
	if platforms != "pc,xbox,ps" {
		platformList := strings.Split(platforms, ",")
		for i := range platformList {
			platformList[i] = strings.TrimSpace(platformList[i])
		}
		err = database.UpdateChannelPlatforms(b, channelID, platformList)
		if err != nil {
			Followup(s, i, fmt.Sprintf("❌ Channel registered but failed to update platforms: %v", err))
			return
		}
	}

	Followup(s, i, fmt.Sprintf("✅ Channel registered for STO news updates!\nPlatforms: %s", platforms))
}

// handleUnregister handles the "unregister" command interaction
func handleUnregister(b *types.Bot, s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Validate inputs
	if i == nil || i.Interaction == nil {
		log.Warning("handleUnregister called with nil interaction")
		return
	}

	channelID := i.ChannelID

	// Check if user has administrator permission
	if !hasAdminPermission(s, i) {
		RespondError(s, i, "You need Administrator permission to use this command.")
		return
	}

	// Remove channel from database
	err := database.RemoveChannel(b, channelID)
	if err != nil {
		log.Errorf("Failed to unregister channel %s: %v", channelID, err)
		RespondError(s, i, "Failed to unregister channel. Please try again later.")
		return
	}

	log.Infof("Channel %s unregistered from STO news", channelID)
	Respond(s, i, "✅ Channel successfully unregistered from Star Trek Online news updates.\n\nThe bot will no longer post news to this channel.")
}

// handleStatus handles the "status" command interaction
func handleStatus(b *types.Bot, s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Validate inputs
	if b == nil {
		log.Warning("handleStatus called with nil bot")
		if s != nil && i != nil {
			RespondError(s, i, "Bot configuration error. Please try again later.")
		}
		return
	}

	if i == nil || i.Interaction == nil {
		log.Warning("handleStatus called with nil interaction")
		return
	}

	channelID := i.ChannelID

	// Check if this channel is registered
	platforms, err := database.GetChannelPlatforms(b, channelID)
	if err != nil {
		log.Errorf("Failed to get channel platforms for %s: %v", channelID, err)
		RespondError(s, i, "Failed to check channel status. Please try again later.")
		return
	}

	// Get cached news count
	allNews, err := database.GetAllCachedNews(b)
	if err != nil {
		log.Errorf("Failed to get cached news count: %v", err)
		RespondError(s, i, "Failed to get bot status. Please try again later.")
		return
	}

	// Build status message
	var statusMsg strings.Builder
	statusMsg.WriteString("🤖 **STOBot Status**\n\n")

	if len(platforms) > 0 {
		statusMsg.WriteString("✅ **This Channel**: Registered\n")
		statusMsg.WriteString(fmt.Sprintf("📡 **Platforms**: %s\n", strings.Join(platforms, ", ")))
	} else {
		statusMsg.WriteString("❌ **This Channel**: Not registered\n")
	}

	statusMsg.WriteString(fmt.Sprintf("📰 **Cached News Items**: %d\n", len(allNews)))
	statusMsg.WriteString(fmt.Sprintf("⏱️ **Poll Period**: %d seconds\n", b.Config.PollPeriod))
	statusMsg.WriteString(fmt.Sprintf("🔔 **Fresh News Threshold**: %d seconds\n", b.Config.FreshSeconds))

	statusMsg.WriteString("\n**Available Commands:**\n")
	statusMsg.WriteString("• `/register` - Register for news updates (Admin only)\n")
	statusMsg.WriteString("• `/unregister` - Unregister from news updates (Admin only)\n")
	statusMsg.WriteString("• `/news` - Get recent news manually\n")
	statusMsg.WriteString("• `/help` - Show help information")

	Respond(s, i, statusMsg.String())
}
