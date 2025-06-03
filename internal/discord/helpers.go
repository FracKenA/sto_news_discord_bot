package discord

import (
	"fmt"
	"strings"

	"github.com/FracKenA/sto_news_discord_bot/internal/types"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

// hasAdminPermission checks if the user has administrator permission in the guild
func hasAdminPermission(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	// If the interaction doesn't have guild info, we can't check permissions
	if i.GuildID == "" || i.Member == nil || i.Member.User == nil {
		log.Debugf("hasAdminPermission: Missing guild info - GuildID: %s, Member: %v, User: %v", i.GuildID, i.Member != nil, i.Member != nil && i.Member.User != nil)
		return false
	}

	log.Debugf("hasAdminPermission: Checking permissions for user %s in guild %s, channel %s", i.Member.User.ID, i.GuildID, i.ChannelID)

	// Try to get user permissions using guild member permissions instead of channel permissions
	guild, err := s.Guild(i.GuildID)
	if err != nil {
		log.Errorf("Failed to get guild info: %v", err)
		return false
	}

	// Check if user is the guild owner
	if i.Member.User.ID == guild.OwnerID {
		log.Debugf("User %s is guild owner", i.Member.User.ID)
		return true
	}

	// Check member roles for administrator permission
	for _, roleID := range i.Member.Roles {
		role, err := s.State.Role(i.GuildID, roleID)
		if err != nil {
			log.Debugf("Failed to get role %s: %v", roleID, err)
			continue
		}

		if role.Permissions&discordgo.PermissionAdministrator != 0 {
			log.Debugf("User %s has administrator permission via role %s", i.Member.User.ID, role.Name)
			return true
		}
	}

	log.Debugf("User %s does not have administrator permission", i.Member.User.ID)
	return false
}

// formatNewsEmbed creates a Discord embed for a news item
func formatNewsEmbed(newsItem types.NewsItem) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       TruncateText(newsItem.Title, 256),
		Description: TruncateText(newsItem.Summary, 2048),
		URL:         fmt.Sprintf("https://playstartrekonline.com/en/news/article/%d", newsItem.ID),
		Color:       0x00ff00, // Green color
		Timestamp:   newsItem.Updated.Format("2006-01-02T15:04:05Z"),
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Platforms: %s", strings.Join(newsItem.Platforms, ", ")),
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Tags",
				Value:  strings.Join(newsItem.Tags, ", "),
				Inline: true,
			},
			{
				Name:   "Platforms",
				Value:  strings.Join(newsItem.Platforms, ", "),
				Inline: true,
			},
		},
	}

	if newsItem.ThumbnailURL != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: newsItem.ThumbnailURL,
		}
	}

	return embed
}
