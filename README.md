# HiuraChat 🗣️

A **cool** **robust** chat bot and framework written in **Go** ✨.

## What's This?

HiuraChat is a WebSocket-based chat bot that's built to be reliable and nice to work with. It's got some neat features like:

- 🔄 Auto-reconnection with smart backoff
- 🚦 Built-in rate limiting to play nice with servers
- 💬 Easy command handling
- 🏓 Ping/pong heartbeat system
- 📝 Clean logging system

## Installation

### Using Pre-built Binary

1. Head over to the [Actions](https://github.com/RuriYS/HiuraChat/actions) page
2. Download the latest artifact for your platform
3. Extract it
4. Make it executable (if you're on Linux):

```bash
chmod +x hiurachat
```

### Using Docker

Clone and run using Docker Compose:

```bash
git clone https://github.com/RuriYS/HiuraChat
cd HiuraChat
docker compose build
docker compose up -d
```

### Building from Source

1. First, make sure you've got Go 1.23+ installed

2. Clone this repo:

```bash
git clone github.com/RuriYS/hiurachat
cd hiurachat
```

3. Build it:

```bash
go build
```

## Configuration

Create a `config.yml` file:

```yaml
bot:
  prefix: "!"          # Command prefix
  response_prefix: ">" # How the bot starts its responses

websocket:
  url: "ws://your-chat-server/ws"

logger:
  level: "info"    # debug, info, warn, or error
  use_colors: true # Pretty colors in console
```

If you're using Docker, mount your config file as shown in the docker-compose.yml:

```yaml
volumes:
  - ./config.yml:/root/config.yml
  - ./logs:/root/logs
```

## Running

### Binary

```bash
./hiurachat
```

### Source

```bash
go run main.go
```

### Docker

```bash
docker compose up -d
```

## Built-in Commands

- `!ping` - Check if the bot is alive (and see the latency!)
- `!echo <message>` - Have the bot repeat something
- `!help <command>` - Get info about commands

## Adding Your Own Commands

It's pretty easy to add new commands! Check out `internal/bot/commands.go` - just add your command to the `initializeCommands` function:

```go
"yourcommand": {
    Name:        "yourcommand",
    Description: "Does something cool",
    Execute: func(args []string) (string, bool) {
        return "> Something cool happened!", true
    },
},
```

## License

This project is licensed under the GNU Affero General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

---

Built with ❤️ and probably too much coffee ☕
