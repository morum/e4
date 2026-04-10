# Contributing

Thanks for contributing to `e4`.

The project is still early, which makes this a good time to help shape both the product and the codebase.

## Ways To Contribute

- report bugs
- improve the terminal UX
- add tests
- improve documentation
- implement focused gameplay or platform features
- review code and raise design concerns early

## Before You Start

For anything non-trivial, open an issue or start a discussion first.

This helps avoid duplicated work and makes it easier to align on:

- product direction
- UX expectations
- scope
- architecture impact

## Development Setup

### Requirements

- Go `1.25+`
- an SSH client

### Run locally

```bash
go run ./cmd/e4 serve --listen :2222 --log-level debug
```

### Test and build

```bash
go test ./...
go build ./...
```

## Project Expectations

When contributing code, aim for changes that are:

- small and focused
- easy to review
- consistent with existing structure
- tested when behavior changes

Prefer explicit, readable code over clever shortcuts.

## Code Guidelines

### General

- preserve existing behavior unless the change is explicitly meant to alter it
- keep the transport layer thin
- keep chess and room logic out of SSH session plumbing where possible
- prefer minimal, direct changes over unnecessary abstraction
- avoid unrelated refactors in the same pull request

### Go-specific

- follow standard Go formatting with `gofmt`
- keep package boundaries clean
- use clear names
- prefer straightforward control flow
- add comments only when they explain non-obvious intent

### Rendering and UX

Changes to the terminal UI should:

- work without ANSI when needed
- degrade cleanly on narrower terminals
- keep important state easy to scan
- avoid noisy redraw behavior

## Pull Requests

### Recommended process

1. create a branch for your change
2. make the smallest correct implementation
3. run `go test ./...`
4. run `go build ./...`
5. update docs if behavior or workflow changed
6. open a pull request with a clear summary

### Good pull requests include

- the problem being solved
- the approach taken
- screenshots or terminal captures for UI changes
- notes about tradeoffs or follow-up work

## Issues

Bug reports are most useful when they include:

- steps to reproduce
- expected behavior
- actual behavior
- relevant logs
- terminal details if the issue is UI-related

## Scope

The project is designed to grow. Contributions that strengthen the foundation are especially valuable, including:

- persistence hooks
- match history
- ratings
- spectator improvements
- room and lobby UX
- operational improvements

## Conduct

Be respectful, direct, and constructive.

Good open source collaboration depends on clear communication and thoughtful review.
