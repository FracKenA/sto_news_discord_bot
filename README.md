# STOBot - Go Edition

A Discord bot for Star Trek Online news, rewritten in Go with SQLite database support and enhanced features.

## Features

- 🚀 **Modern Go Implementation**: Fast, reliable, and maintainable
- 📦 **SQLite Database**: Persistent storage for channels, news cache, and posted items
- 🔧 **Slash Commands**: Modern Discord interactions with user-friendly commands
- 🎯 **Platform Filtering**: Filter news by PC, Xbox, and PlayStation platforms
- ⚡ **Real-time Polling**: Automated news fetching and posting
- 🛡️ **Duplicate Detection**: Intelligent duplicate checking in recent messages
- 📊 **Health Monitoring**: Built-in health checks and logging
- 🐳 **Docker Support**: Ready-to-deploy containerized application

## Quick Start

### Using Docker Compose (Recommended)

1. **Clone and setup**:
   ```bash
   cd /home/kdobbins/compose/stobot
   cp .env.example .env
   ```

2. **Configure environment**:
   Edit `.env` and set your Discord bot token:
   ```
   DISCORD_TOKEN=your_discord_bot_token_here
   ```

3. **Start the Go version**:
   ```bash
   docker-compose up -d stobot
   ```

4. **View logs**:
   ```bash
   docker-compose logs -f stobot
   ```

### Using the Rust version (legacy)

```bash
docker-compose --profile rust up -d stobot-rust
```

## Slash Commands

### Admin Commands (requires Administrator permission)
- `/stobot_register` - Register this channel for STO news
- `/stobot_unregister` - Unregister this channel from STO news  
- `/stobot_status` - Show current bot configuration

### General Commands
- `/stobot_news [platforms] [weeks]` - Show recent STO news
- `/stobot_patchnotes [platforms] [weeks]` - Show recent patch notes
- `/stobot_help` - Show available commands

### Command Examples

```
/stobot_news platforms:pc,xbox weeks:2
/stobot_patchnotes platforms:pc weeks:1
/stobot_register
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DISCORD_TOKEN` | *required* | Discord bot token |
| `POLL_PERIOD` | `600` | Seconds between news checks |
| `POLL_COUNT` | `20` | Number of news items to fetch |
| `FRESH_SECONDS` | `600` | Max age of news to post (seconds) |
| `MSG_COUNT` | `10` | Messages to check for duplicates |
| `CHANNELS_PATH` | `/data/channels.txt` | Path to channels file |
| `DATABASE_PATH` | `/data/stobot.db` | Path to SQLite database |

### Command Line Options

```bash
stobot --help
stobot --poll-period 300 --fresh-seconds 1200
```

## Database Schema

The bot uses SQLite with the following tables:

- **channels**: Registered Discord channels and their platform preferences
- **posted_news**: Track which news items have been posted to prevent duplicates
- **news_cache**: Cache fetched news for performance and offline access

## Development

### Prerequisites

- Go 1.21+
- SQLite3
- Discord bot token

### Local Development

1. **Install dependencies**:
   ```bash
   go mod download
   ```

2. **Build the application**:
   ```bash
   CGO_ENABLED=1 go build -o stobot .
   ```

3. **Run locally**:
   ```bash
   export DISCORD_TOKEN=your_token_here
   ./stobot --database-path ./stobot.db
   ```

### CLI Commands

The bot includes several command-line utilities for management and maintenance:

#### Database Population
```bash
# Populate database with historical news (prevents re-posting old articles)
./stobot populate-db --count 100 --tags star-trek-online,patch-notes

# Dry run to see what would be populated
./stobot populate-db --dry-run --count 50
```

#### Channel Management
```bash
# Import channels from legacy channels.txt format
./stobot import-channels --channels-file ./channels.txt

# List all registered channels
./stobot list-channels
```

#### News Management
```bash
# Mark all cached news as posted (prevents re-sending existing news)
./stobot mark-all-posted

# Dry run to see what would be marked
./stobot mark-all-posted --dry-run
```

#### Command Options
All commands support:
- `--database-path` - Path to SQLite database (default: `./data/stobot.db`)
- `--help` - Show detailed help for each command

### Building Docker Image

```bash
docker build -f Dockerfile.go -t stobot:go .
```

## Migration from Rust Version

The Go version is designed as a drop-in replacement for the Rust version:

1. **Stop the Rust version**:
   ```bash
   docker-compose down stobot-rust
   ```

2. **Start the Go version**:
   ```bash
   docker-compose up -d stobot
   ```

3. **Existing channel registrations** will be migrated automatically from `channels.txt`

## Features Comparison

| Feature | Rust Version | Go Version |
|---------|-------------|------------|
| Basic news posting | ✅ | ✅ |
| Slash commands | ✅ | ✅ |
| Platform filtering | ✅ | ✅ |
| SQLite database | ❌ | ✅ |
| News caching | ❌ | ✅ |
| Enhanced logging | ✅ | ✅ |
| Health checks | ✅ | ✅ |
| AI readiness | ❌ | ✅ (prepared) |

## Logging

The bot uses structured JSON logging with different levels:

- **INFO**: General operational messages
- **ERROR**: Error conditions that need attention
- **DEBUG**: Detailed debugging information

Example log entry:
```json
{
  "level": "info",
  "msg": "Posted news item 12345 ('New Ship Release') to channel 987654321",
  "time": "2025-05-26T10:30:00Z"
}
```

## Health Monitoring

The bot includes health checks accessible via:
- Docker health checks
- Process monitoring
- Database connectivity checks

## Troubleshooting

### Common Issues

1. **Bot not responding to commands**:
   - Check Discord token is correct
   - Verify bot has necessary permissions
   - Check logs for authentication errors

2. **News not posting**:
   - Verify channels are registered with `/stobot_status`
   - Check API connectivity
   - Review fresh_seconds setting

3. **Database errors**:
   - Ensure /data directory is writable
   - Check SQLite file permissions
   - Review database path configuration

### Debug Mode

Enable debug logging:
```bash
docker-compose up -d stobot
docker-compose exec stobot sh -c 'export LOG_LEVEL=debug'
```

## Future Enhancements

The Go version is prepared for:
- 🤖 **AI Integration**: ChatGPT/Claude API for news summarization
- 📈 **Analytics**: Usage statistics and trending topics  
- 🔔 **Advanced Notifications**: Webhook support and custom filters
- 🌐 **Web Dashboard**: Web interface for configuration
- 📱 **Mobile Apps**: REST API for mobile clients

## Legal and Compliance

STOBot follows Discord's developer policies and best practices:

- **Privacy Policy**: See [PRIVACY_POLICY.md](PRIVACY_POLICY.md) for data handling practices
- **Terms of Service**: See [TERMS_OF_SERVICE.md](TERMS_OF_SERVICE.md) for usage terms
- **Bot Verification**: See [VERIFICATION_GUIDE.md](VERIFICATION_GUIDE.md) for verification readiness

## License

See [LICENSE](LICENSE) file for details.

## Support

For issues and questions:
1. Check the logs: `docker-compose logs stobot`
2. Review this documentation
3. Check Discord bot permissions
4. Verify environment configuration

---

**Star Trek Online** and related marks are trademarks of CBS Studios Inc. This bot is not affiliated with or endorsed by CBS Studios Inc. or Perfect World Entertainment.
