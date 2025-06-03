package discord

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/FracKenA/sto_news_discord_bot/internal/types"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

// GameStatusResponse represents the structure of the STO server status API response
type GameStatusResponse struct {
	ServerStatus string `json:"server_status"`
}

// handleGameStatus handles the "game_status" command interaction
func handleGameStatus(b *types.Bot, s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Acknowledge interaction with timeout handling
	if err := AcknowledgeWithRetry(s, i); err != nil {
		log.Errorf("Failed to acknowledge game_status command: %v", err)
		return
	}

	// Fetch server status from STO launcher API
	log.Info("Fetching STO server status from launcher API")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get("http://launcher.startrekonline.com/launcher_server_status")
	if err != nil {
		log.Errorf("Failed to fetch server status: %v", err)
		Followup(s, i, "❌ Failed to fetch server status. Please try again later.")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Errorf("Server status API returned status %d", resp.StatusCode)
		Followup(s, i, "❌ Server status API is currently unavailable.")
		return
	}

	var statusResponse GameStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&statusResponse); err != nil {
		log.Errorf("Failed to decode server status response: %v", err)
		Followup(s, i, "❌ Failed to parse server status response.")
		return
	}

	// Create embed for server status
	statusEmoji := "✅"
	statusColor := 0x00ff00 // Green

	switch strings.ToUpper(statusResponse.ServerStatus) {
	case "UP":
		statusEmoji = "✅"
		statusColor = 0x00ff00 // Green
	case "DOWN":
		statusEmoji = "❌"
		statusColor = 0xff0000 // Red
	case "MAINTENANCE":
		statusEmoji = "🔧"
		statusColor = 0xffaa00 // Orange
	default:
		statusEmoji = "❓"
		statusColor = 0x808080 // Gray
	}

	embed := &discordgo.MessageEmbed{
		Title:       "🎮 Star Trek Online Server Status",
		Description: fmt.Sprintf("%s **%s**", statusEmoji, strings.ToUpper(statusResponse.ServerStatus)),
		Color:       statusColor,
		Timestamp:   time.Now().Format("2006-01-02T15:04:05Z"),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Data from launcher.startrekonline.com",
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Status",
				Value:  strings.ToUpper(statusResponse.ServerStatus),
				Inline: true,
			},
			{
				Name:   "Last Checked",
				Value:  time.Now().Format("15:04:05 UTC"),
				Inline: true,
			},
		},
	}

	// Send the result with enhanced error handling
	if err := FollowupWithEmbeds(s, i, "", []*discordgo.MessageEmbed{embed}); err != nil {
		log.Errorf("Failed to send server status: %v", err)
		Followup(s, i, "❌ Failed to send server status.")
		return
	}

	log.Infof("Sent STO server status: %s", statusResponse.ServerStatus)
}
