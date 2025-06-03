# Privacy Policy for STOBot

**Effective Date**: December 18, 2024  
**Last Updated**: December 18, 2024

## Introduction

STOBot ("the Bot") is a Discord bot that provides Star Trek Online news updates to registered Discord channels. This Privacy Policy explains how we collect, use, and protect information when you use our bot.

## Information We Collect

### Discord Data
We collect only the minimal data necessary to provide our service:

- **Channel IDs**: Discord channel identifiers where the bot is registered
- **Platform Preferences**: User-selected gaming platforms (PC, Xbox, PlayStation) for news filtering
- **Interaction Data**: Basic interaction logs for debugging and service improvement

### News Data
- **Cached News Articles**: We cache Star Trek Online news from official APIs to improve performance
- **Posted News Tracking**: We track which news items have been posted to prevent duplicates

## How We Use Your Information

We use the collected information solely to:

1. **Deliver News Updates**: Post Star Trek Online news to registered Discord channels
2. **Prevent Duplicates**: Track posted news to avoid sending duplicate content
3. **Filter Content**: Respect platform preferences for relevant news delivery
4. **Improve Service**: Basic logging for debugging and service reliability

## Data Storage and Security

- **Local Storage**: All data is stored locally in an SQLite database
- **No Third-Party Sharing**: We do not share any data with third parties
- **Data Retention**: Channel registrations persist until manually removed
- **News Cache**: News articles are cached temporarily for performance

## Data Access and Control

You have the following rights regarding your data:

- **Registration Control**: Server administrators can register/unregister channels using `/stobot_register` and `/stobot_unregister`
- **Data Viewing**: Use `/stobot_status` to view current channel registration status
- **Data Removal**: Unregistering a channel removes all associated data

## Data We Do NOT Collect

We explicitly do not collect:

- Personal user information or profiles
- Message content from your Discord channels
- User IDs or personal identifiers
- Voice or video data
- Financial or payment information

## Third-Party Services

The bot fetches news from:
- **Arc Games API**: Official Star Trek Online news source
- **Discord API**: For bot functionality only

We do not control these third-party services' privacy practices.

## Children's Privacy

Our bot does not knowingly collect information from users under 13 years of age.

## Changes to This Policy

We may update this Privacy Policy occasionally. Significant changes will be announced in bot updates or through our support channels.

## Contact Information

For privacy-related questions or concerns:

- **Support**: Check bot logs using `docker-compose logs stobot`
- **Issues**: Review documentation and Discord bot permissions
- **Data Requests**: Use bot commands to manage your channel registrations

## Compliance

This bot operates in compliance with:
- Discord's Developer Terms of Service
- Discord's Developer Policy
- General data protection principles

---

**Note**: Star Trek Online and related marks are trademarks of CBS Studios Inc. This bot is not affiliated with or endorsed by CBS Studios Inc. or Perfect World Entertainment.
