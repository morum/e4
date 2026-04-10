# e4

`e4` is a terminal-native chess server you join over SSH.

It is built for the feeling of old-school network play: fast login, clean text UX, no browser required, and games that feel at home inside a shell.

## Why e4

- SSH-first multiplayer chess
- smooth terminal UX with ANSI-enhanced boards and status views
- create, join, and watch live games
- standard time controls like `10|0`, `3|2`, and `15|10`
- SAN move input like `e4`, `Nf3`, `O-O`, and `Qxe5+`
- modular Go codebase designed to grow into a full platform

## Current Status

`e4` is already playable and actively evolving.

Today’s foundation includes:

- a single-binary Go server
- SSH session handling with branded join flow
- lobby and room lifecycle management
- live board rendering with piece glyphs and ANSI styling
- spectator mode
- chess clocks with real-time updates
- structured logging for debugging and operations

The project is intentionally early, but the architecture is aimed at long-term expansion: persistence, ratings, chat, rematches, tournaments, bots, and additional clients can be added without rewriting the chess core.

## Quick Start

### Run locally

```bash
go run ./cmd/e4 serve --listen :2222 --log-level debug
```

### Connect

```bash
ssh -p 2222 anything@localhost
```

The SSH username is ignored in the current version. You choose a nickname after connecting.

### Install the binary

```bash
go install github.com/morum/e4/cmd/e4@latest
```

Then run:

```bash
e4 serve
```

## What It Feels Like

Once connected, players land in a lobby where they can:

- `create <tc>` to open a game
- `join <id>` to take a seat
- `watch <id>` to spectate
- `list` to refresh the lobby

Inside a room, players see:

- a live board
- clocks for both sides
- move history
- room status and watcher count
- contextual prompts like `white move>` and `watch[ROOM]>`

Moves are entered in standard algebraic notation:

```text
e4
Nf3
O-O
Qxe5+
```

## Commands

### Lobby

```text
list
create 10|0
join ABC123
watch ABC123
help
quit
```

### Room

```text
e4
Nf3
O-O
board
leave
resign
help
quit
```

## Configuration

The main server entrypoint is:

```bash
e4 serve [--listen :2222] [--host-key ./.e4_host_key] [--log-level info]
```

### Flags

- `--listen`: SSH bind address
- `--host-key`: path to the SSH private host key file
- `--log-level`: `debug`, `info`, `warn`, or `error`

By default, `e4` stores its generated host key in `.e4_host_key`.

## Architecture

The codebase is split so the transport layer stays thin and the game logic remains reusable.

```text
cmd/e4                  CLI entrypoint
internal/app            app wiring and configuration
internal/domain         pure game and lobby types
internal/service        room and lobby services
internal/store/memory   in-memory repository implementations
internal/clock          chess clock state
internal/render         terminal rendering and theming
internal/transport/ssh  SSH transport and session UX
```

This layout is meant to support future work such as:

- persistent accounts
- game history and PGN export
- ratings
- rematches and draw offers
- chat
- bots
- tournaments
- alternate frontends

## Development

### Requirements

- Go `1.25+`
- an SSH client

### Common commands

```bash
go test ./...
go build ./...
go run ./cmd/e4 serve --listen :2222 --log-level debug
```

## Roadmap

The next major layers for `e4` are:

- persistent storage and player identity
- stronger multiplayer flows like rematch and draw offers
- richer spectator experience
- ratings and matchmaking
- tournament orchestration
- bots and analysis tools

## Contributing

Contributions are welcome.

Start here:

- read [`CONTRIBUTING.md`](./CONTRIBUTING.md)
- open an issue for bugs, UX problems, or feature ideas
- send a pull request when you have a focused improvement ready

## License

`e4` is released under the [MIT License](./LICENSE).
