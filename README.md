# e4

`e4` is a chess server for the terminal.

You run the server, connect over SSH, pick a nickname, and play or watch games.

![e4 screenshot](./screenshot.png)

## What It Does

- serves chess over SSH
- lets players create, join, and watch games
- supports time controls like `10|0`, `3|2`, and `15|10`
- accepts SAN moves like `e4`, `Nf3`, `O-O`, and `Qxe5+`
- shows a live board, clocks, and move list in the terminal
- supports tab completion for room IDs in `join` and `watch`

## Quick Start

Run the server:

```bash
go run ./cmd/e4 serve --listen :2222 --log-level debug
```

Connect from another terminal:

```bash
ssh -p 2222 anything@localhost
```

The SSH username is ignored. You choose a nickname after connecting.

Install the binary with:

```bash
go install github.com/morum/e4/cmd/e4@latest
```

Then run:

```bash
e4 serve
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

Tips:

- press `Tab` after `join ` or `watch ` to autocomplete room IDs
- press `Tab` again on the same partial input to list matching room IDs

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

```bash
e4 serve [--listen :2222] [--host-key ./.e4_host_key] [--log-level info]
```

Flags:

- `--listen`: SSH bind address
- `--host-key`: path to the SSH private host key file
- `--log-level`: `debug`, `info`, `warn`, or `error`

By default, `e4` stores its generated host key in `.e4_host_key`.

## Project Layout

```text
cmd/e4                  CLI entrypoint
internal/app            app wiring and configuration
internal/domain         core game and lobby types
internal/service        room and lobby services
internal/store/memory   in-memory repositories
internal/clock          chess clock state
internal/render         terminal rendering and theming
internal/transport/ssh  SSH transport and session handling
```

The code is structured so persistence, ratings, chat, bots, tournaments, and other clients can be added later without replacing the core game flow.

## Development

Requirements:

- Go `1.25+`
- an SSH client

Common commands:

```bash
go test ./...
go build ./...
go run ./cmd/e4 serve --listen :2222 --log-level debug
```

## Contributing

See [`CONTRIBUTING.md`](./CONTRIBUTING.md).

## License

`e4` is released under the [MIT License](./LICENSE).
