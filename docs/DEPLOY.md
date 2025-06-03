# Stobot Deployment Guide

## Quick Start

1. **Copy environment file**:
   ```bash
   cp .env.example .env
   ```

2. **Configure your Discord token**:
   Edit `.env` and set your Discord bot token:
   ```
   DISCORD_TOKEN=your_actual_discord_bot_token_here
   ```

3. **Start the application**:
   ```bash
   docker-compose up -d
   ```

## Configuration

### Environment Variables

- `DISCORD_TOKEN` - Required Discord bot token
- `POLL_PERIOD` - Seconds between news checks (default: 600)
- `POLL_COUNT` - Number of news items to fetch per poll (default: 20)
- `FRESH_SECONDS` - Maximum age of news items to post (default: 600)
- `MSG_COUNT` - Discord messages to check for duplicates (default: 10)
- `CHANNELS_PATH` - Path to channels file (default: /data/channels.txt)
- `RUST_LOG` - Logging level (error, warn, info, debug, trace) - *Note: Still uses RUST_LOG for compatibility*

### Volume Mounts

- `./channels.txt:/data/channels.txt:ro` - Channel configuration (read-only)
- `./data:/data` - Persistent data storage

## Security Features

- Non-root user execution in container
- Read-only channel configuration
- Environment variable configuration
- Health checks enabled

## Monitoring

The container includes health checks that verify the application is running.
Check status with:
```bash
docker-compose ps
```

## Logs

View application logs:
```bash
docker-compose logs -f stobot
```

## Updating

1. Pull new image:
   ```bash
   docker-compose pull
   ```

2. Restart service:
   ```bash
   docker-compose up -d
   ```

## Migration from Legacy Version

### Importing Existing Channels

If you're upgrading from the legacy Rust version, you can import your existing `channels.txt` file into the new database:

1. **Build the Go version**:
   ```bash
   docker-compose build
   ```

2. **Import channels**:
   ```bash
   docker-compose run --rm stobot import-channels --channels-file /data/channels.txt --database-path /data/stobot.db
   ```

3. **Verify import**:
   ```bash
   docker-compose run --rm stobot populate-db --dry-run
   ```

The import process will:
- Parse the existing `channels.txt` format (`channel:ID|platforms`)
- Import channels into the SQLite database
- Skip channels that already exist
- Validate channel IDs and platform configurations

### Populating Historical News

To prevent re-posting old news articles when starting the bot:

```bash
docker-compose run --rm stobot populate-db --count 100 --tags star-trek-online,patch-notes
```

This will fetch recent news and mark it as already posted.

### Marking Existing News as Posted

If you already have cached news in the database but want to mark it all as posted to prevent re-sending:

```bash
# Check what would be marked (dry run)
docker-compose run --rm stobot mark-all-posted --dry-run

# Actually mark all cached news as posted
docker-compose run --rm stobot mark-all-posted
```

This is useful when:
- Migrating from the Rust version with existing news cache
- Recovering from a database issue where posted status was lost
- Starting fresh without re-posting historical content

### Database Migration

The Go version includes automatic database migrations that run when the bot starts:

- **Posted News Schema**: Automatically migrates from the old schema (where `news_id` was PRIMARY KEY) to the new schema that allows the same news to be posted to multiple channels
- **Tags Column**: Adds the `tags` column to the `news_cache` table if missing
- **Indexes**: Creates necessary database indexes for performance

No manual intervention is required - migrations run automatically on startup.
