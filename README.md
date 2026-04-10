# chessh

`chessh` is a terminal-first chess server served over SSH.

The first cut is intentionally small but modular:

- Go single-binary server
- `chessh serve` entrypoint
- in-memory lobby and game rooms
- create, join, and watch flows
- SAN move input like `e4`, `Nf3`, `O-O`, `Qxe5+`
- major time controls like `10|0`, `3|2`, `15|10`

## Run

```bash
go run ./cmd/chessh serve --listen :2222 --log-level debug
```

Then connect from another terminal:

```bash
ssh -p 2222 anything@localhost
```

The SSH username is ignored in v1. You choose a nickname after connecting.

Server logs are written to stderr with the standard Go `slog` text format.

## Lobby Commands

```text
list
create 10|0
join ABC123
watch ABC123
help
quit
```

## Room Commands

```text
e4
Nf3
O-O
leave
resign
board
help
```

## Project Layout

```text
cmd/chessh              CLI entrypoint
internal/app            app wiring and configuration
internal/domain         pure game and lobby types
internal/service        room and lobby services
internal/store/memory   in-memory repository implementations
internal/clock          chess clock state
internal/render         text rendering
internal/transport/ssh  SSH transport and session UX
```

The code is structured so persistence, ratings, accounts, rematches, chat, and additional transports can be added without rewriting the chess core.
