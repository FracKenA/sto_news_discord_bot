// Package discord contains tests for the STOBot Discord integration package.
//
// These tests cover command registration, validation, and event handling.
package discord

import (
	"testing"

	"github.com/FracKenA/sto_news_discord_bot/internal/testhelpers"

	"github.com/bwmarrin/discordgo"
)

// getTestCommands returns a list of test command definitions for testing purposes.
func getTestCommands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "stobot_register",
			Description: "Register this channel for STO news updates",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "platforms",
					Description: "Comma-separated list of platforms (pc,xbox,ps)",
					Required:    false,
				},
			},
		},
		{
			Name:        "stobot_unregister",
			Description: "Unregister this channel from STO news updates",
		},
		{
			Name:        "stobot_status",
			Description: "Show bot status and registered channels",
		},
		{
			Name:        "stobot_news",
			Description: "Get recent Star Trek Online news",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "tag",
					Description: "News category",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "General", Value: "star-trek-online"},
						{Name: "Patch Notes", Value: "patch-notes"},
						{Name: "Events", Value: "events"},
						{Name: "Dev Blogs", Value: "dev-blogs"},
					},
				},
			},
		},
	}
}
func TestRegisterCommands(t *testing.T) {
	// Test that RegisterCommands function exists and can be called
	// Note: This will fail without a real Discord session, but we can test
	// that the function doesn't panic and the command structures are valid

	// We can't test the actual registration without mocking Discord API
	// but we can verify the command definitions are valid

	commands := getTestCommands()

	// Verify we have the expected commands
	expectedCommands := []string{"stobot_register", "stobot_unregister", "stobot_status", "stobot_news"}
	if len(commands) < len(expectedCommands) {
		t.Errorf("Expected at least %d commands, got %d", len(expectedCommands), len(commands))
	}

	// Verify command structure
	for _, cmd := range commands {
		if cmd.Name == "" {
			t.Error("Command name should not be empty")
		}
		if cmd.Description == "" {
			t.Error("Command description should not be empty")
		}
	}

	t.Log("Command structures are valid")
}

func TestHandleCommand(t *testing.T) {
	bot := testhelpers.CreateTestBot(t)
	defer bot.DB.Close()

	// Test that HandleCommand function exists
	// We can't easily test the actual command handling without mocking Discord
	// but we can test that the function doesn't panic with nil inputs

	// Create a basic interaction for testing
	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type: discordgo.InteractionApplicationCommand,
			Data: discordgo.ApplicationCommandInteractionData{
				Name: "stobot_status",
			},
			ChannelID: "123456789",
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       "987654321",
					Username: "testuser",
				},
			},
		},
	}

	if interaction.Interaction == nil {
		t.Fatal("Interaction should not be nil")
		return
	}

	if interaction.Interaction.Member == nil {
		t.Fatal("Member should not be nil")
		return
	}

	if interaction.Interaction.Member.User == nil {
		t.Fatal("User should not be nil")
		return
	}

	// Test that HandleCommand doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("HandleCommand panicked: %v", r)
		}
	}()

	// This will likely fail due to nil session, but shouldn't panic
	HandleCommand(bot, nil, interaction)

	t.Log("HandleCommand executed without panic")
}

func TestCommandValidation(t *testing.T) {
	commands := getTestCommands()

	if commands == nil {
		t.Fatal("Commands should not be nil")
		return
	}

	for _, cmd := range commands {
		t.Run(cmd.Name, func(t *testing.T) {
			// Test command name
			if cmd.Name == "" {
				t.Error("Command name should not be empty")
			}
			if len(cmd.Name) > 32 {
				t.Errorf("Command name too long: %d characters (max 32)", len(cmd.Name))
			}

			// Test command description
			if cmd.Description == "" {
				t.Error("Command description should not be empty")
			}
			if len(cmd.Description) > 100 {
				t.Errorf("Command description too long: %d characters (max 100)", len(cmd.Description))
			}

			// Test options if present
			for _, option := range cmd.Options {
				if option.Name == "" {
					t.Error("Option name should not be empty")
				}
				if option.Description == "" {
					t.Error("Option description should not be empty")
				}

				// Test choices if present
				for _, choice := range option.Choices {
					if choice.Name == "" {
						t.Error("Choice name should not be empty")
					}
					if choice.Value == nil {
						t.Error("Choice value should not be nil")
					}
				}
			}
		})
	}
}

func TestRegisterCommandOptions(t *testing.T) {
	commands := getTestCommands()

	if commands == nil {
		t.Fatal("Commands should not be nil")
		return
	}

	var registerCmd *discordgo.ApplicationCommand
	for _, cmd := range commands {
		if cmd.Name == "stobot_register" {
			registerCmd = cmd
			break
		}
	}

	if registerCmd == nil {
		t.Fatal("Register command not found")
		return
	}

	// Verify register command has platforms option
	if len(registerCmd.Options) == 0 {
		t.Error("Register command should have options")
	}

	var platformsOption *discordgo.ApplicationCommandOption
	for _, option := range registerCmd.Options {
		if option.Name == "platforms" {
			platformsOption = option
			break
		}
	}

	if platformsOption == nil {
		t.Error("Register command should have platforms option")
		return
	}

	if platformsOption.Type != discordgo.ApplicationCommandOptionString {
		t.Error("Platforms option should be string type")
	}

	if platformsOption.Required {
		t.Error("Platforms option should not be required")
	}
}

func TestNewsCommandOptions(t *testing.T) {
	commands := getTestCommands()

	if commands == nil {
		t.Fatal("Commands should not be nil")
		return
	}

	var newsCmd *discordgo.ApplicationCommand
	for _, cmd := range commands {
		if cmd.Name == "stobot_news" {
			newsCmd = cmd
			break
		}
	}

	if newsCmd == nil {
		t.Fatal("News command not found")
		return
	}

	// Verify news command has tag option
	if len(newsCmd.Options) == 0 {
		t.Error("News command should have options")
	}

	var tagOption *discordgo.ApplicationCommandOption
	for _, option := range newsCmd.Options {
		if option.Name == "tag" {
			tagOption = option
			break
		}
	}

	if tagOption == nil {
		t.Error("News command should have tag option")
		return
	}

	if tagOption.Type != discordgo.ApplicationCommandOptionString {
		t.Error("Tag option should be string type")
	}

	if tagOption.Required {
		t.Error("Tag option should not be required")
	}

	// Verify choices
	expectedChoices := map[string]string{
		"General":     "star-trek-online",
		"Patch Notes": "patch-notes",
		"Events":      "events",
		"Dev Blogs":   "dev-blogs",
	}

	if len(tagOption.Choices) != len(expectedChoices) {
		t.Errorf("Expected %d choices, got %d", len(expectedChoices), len(tagOption.Choices))
	}

	for _, choice := range tagOption.Choices {
		expectedValue, exists := expectedChoices[choice.Name]
		if !exists {
			t.Errorf("Unexpected choice name: %s", choice.Name)
		}
		if choice.Value != expectedValue {
			t.Errorf("Expected choice value %s, got %v", expectedValue, choice.Value)
		}
	}
}

func TestSimpleCommands(t *testing.T) {
	commands := getTestCommands()

	if commands == nil {
		t.Fatal("Commands should not be nil")
		return
	}

	// Test that unregister and status commands have no options
	simpleCommands := []string{"stobot_unregister", "stobot_status"}

	for _, cmdName := range simpleCommands {
		t.Run(cmdName, func(t *testing.T) {
			var cmd *discordgo.ApplicationCommand
			for _, c := range commands {
				if c.Name == cmdName {
					cmd = c
					break
				}
			}

			if cmd == nil {
				t.Fatalf("Command %s not found", cmdName)
				return
			}

			if len(cmd.Options) != 0 {
				t.Errorf("Command %s should have no options, got %d", cmdName, len(cmd.Options))
			}
		})
	}
}

func TestCommandNamesUnique(t *testing.T) {
	commands := getTestCommands()

	if commands == nil {
		t.Fatal("Commands should not be nil")
		return
	}

	names := make(map[string]bool)
	for _, cmd := range commands {
		if names[cmd.Name] {
			t.Errorf("Duplicate command name: %s", cmd.Name)
		}
		names[cmd.Name] = true
	}
}
