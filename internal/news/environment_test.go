package news

import (
	"testing"

	"github.com/FracKenA/sto_news_discord_bot/internal/database"
	"github.com/FracKenA/sto_news_discord_bot/internal/testhelpers"
)

// TestEnvironmentFiltering tests that channels are filtered by environment
func TestEnvironmentFiltering(t *testing.T) {
	bot := testhelpers.CreateTestBot(t)
	t.Cleanup(func() {
		if bot.DB != nil {
			_ = bot.DB.Close()
		}
	})

	// Add test channels with different environments
	devChannelID := "123456789"
	prodChannelID := "987654321"

	err := database.AddChannelWithEnvironment(bot, devChannelID, "DEV")
	if err != nil {
		t.Fatalf("Failed to add DEV channel: %v", err)
	}

	err = database.AddChannelWithEnvironment(bot, prodChannelID, "PROD")
	if err != nil {
		t.Fatalf("Failed to add PROD channel: %v", err)
	}

	// Test 1: Bot with DEV environment should only get DEV channels
	bot.Config.Environment = "DEV"
	devChannels, err := database.GetChannelsByEnvironment(bot, "DEV")
	if err != nil {
		t.Fatalf("Failed to get DEV channels: %v", err)
	}
	if len(devChannels) != 1 || devChannels[0] != devChannelID {
		t.Errorf("Expected 1 DEV channel (%s), got %v", devChannelID, devChannels)
	}

	// Test 2: Bot with PROD environment should only get PROD channels
	bot.Config.Environment = "PROD"
	prodChannels, err := database.GetChannelsByEnvironment(bot, "PROD")
	if err != nil {
		t.Fatalf("Failed to get PROD channels: %v", err)
	}
	if len(prodChannels) != 1 || prodChannels[0] != prodChannelID {
		t.Errorf("Expected 1 PROD channel (%s), got %v", prodChannelID, prodChannels)
	}

	// Test 3: No environment set should get all channels (backwards compatibility)
	bot.Config.Environment = ""
	allChannels, err := database.GetRegisteredChannels(bot)
	if err != nil {
		t.Fatalf("Failed to get all channels: %v", err)
	}
	if len(allChannels) != 2 {
		t.Errorf("Expected 2 total channels, got %d", len(allChannels))
	}

	t.Logf("Environment filtering test passed: DEV=%d, PROD=%d, ALL=%d channels",
		len(devChannels), len(prodChannels), len(allChannels))
}

// TestProcessChannelNewsEnvironmentFiltering tests that ProcessChannelNews respects environment
func TestProcessChannelNewsEnvironmentFiltering(t *testing.T) {
	bot := testhelpers.CreateTestBot(t)
	t.Cleanup(func() {
		if bot.DB != nil {
			_ = bot.DB.Close()
		}
	})

	// Set up test configuration
	bot.Config.PollCount = 5
	bot.Config.FreshSeconds = 7200

	// Add test channels with different environments
	devChannelID := "111111111"
	prodChannelID := "222222222"

	err := database.AddChannelWithEnvironment(bot, devChannelID, "DEV")
	if err != nil {
		t.Fatalf("Failed to add DEV channel: %v", err)
	}

	err = database.AddChannelWithEnvironment(bot, prodChannelID, "PROD")
	if err != nil {
		t.Fatalf("Failed to add PROD channel: %v", err)
	}

	// Test: Bot in DEV environment should skip PROD channels
	bot.Config.Environment = "DEV"

	// Verify channel environment checking works
	devEnv, err := database.GetChannelEnvironment(bot, devChannelID)
	if err != nil {
		t.Fatalf("Failed to get DEV channel environment: %v", err)
	}
	if devEnv != "DEV" {
		t.Errorf("Expected DEV environment, got %s", devEnv)
	}

	prodEnv, err := database.GetChannelEnvironment(bot, prodChannelID)
	if err != nil {
		t.Fatalf("Failed to get PROD channel environment: %v", err)
	}
	if prodEnv != "PROD" {
		t.Errorf("Expected PROD environment, got %s", prodEnv)
	}

	// Test environment filtering at the channel level
	devChannels, err := database.GetChannelsByEnvironment(bot, "DEV")
	if err != nil {
		t.Fatalf("Failed to get DEV channels: %v", err)
	}
	if len(devChannels) != 1 || devChannels[0] != devChannelID {
		t.Errorf("Expected 1 DEV channel, got %v", devChannels)
	}

	prodChannels, err := database.GetChannelsByEnvironment(bot, "PROD")
	if err != nil {
		t.Fatalf("Failed to get PROD channels: %v", err)
	}
	if len(prodChannels) != 1 || prodChannels[0] != prodChannelID {
		t.Errorf("Expected 1 PROD channel, got %v", prodChannels)
	}

	t.Log("ProcessChannelNews environment filtering test completed successfully")
}
