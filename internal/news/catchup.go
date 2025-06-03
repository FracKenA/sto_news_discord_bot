package news

import (
	"time"

	"github.com/FracKenA/sto_news_discord_bot/internal/database"
	"github.com/FracKenA/sto_news_discord_bot/internal/types"
	log "github.com/sirupsen/logrus"
)

// CatchUpUnpostedNews posts any unposted news items from the last N days to all registered channels.
func CatchUpUnpostedNews(b *types.Bot, days int) {
	channels, err := database.GetRegisteredChannels(b)
	if err != nil {
		log.Errorf("[catchup] Failed to get registered channels: %v", err)
		return
	}
	if len(channels) == 0 {
		log.Info("[catchup] No registered channels found, skipping catch-up.")
		return
	}

	tags := []string{"star-trek-online", "patch-notes"}
	cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour)

	for _, tag := range tags {
		newsItems, err := FetchNews(b, tag, b.Config.PollCount*10, BulkFetchOptions())
		if err != nil {
			log.Errorf("[catchup] Failed to fetch news for tag %s: %v", tag, err)
			continue
		}
		for _, channelID := range channels {
			platforms, err := database.GetChannelPlatforms(b, channelID)
			if err != nil {
				log.Errorf("[catchup] Failed to get platforms for channel %s: %v", channelID, err)
				continue
			}
			filteredNews := filterNewsByPlatforms(newsItems, platforms)
			for _, newsItem := range filteredNews {
				if newsItem.Updated.Before(cutoff) {
					continue
				}
				posted, err := database.IsNewsPosted(b, newsItem.ID, channelID)
				if err != nil {
					log.Errorf("[catchup] Failed to check posted for news %d: %v", newsItem.ID, err)
					continue
				}
				if posted {
					continue
				}
				if IsDuplicateInRecentMessages(b, channelID, newsItem) {
					continue
				}
				if err := PostNewsToChannel(b, channelID, newsItem); err != nil {
					log.Errorf("[catchup] Failed to post news %d to channel %s: %v", newsItem.ID, channelID, err)
					continue
				}
				if err := database.MarkNewsAsPosted(b, newsItem.ID, channelID); err != nil {
					log.Errorf("[catchup] Failed to mark news %d as posted: %v", newsItem.ID, err)
				}
				log.Infof("[catchup] Posted news item %d ('%s') to channel %s", newsItem.ID, newsItem.Title, channelID)
			}
		}
	}
}
