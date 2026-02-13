# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

CLE (Command-Line Editor) is a Go library that provides interactive command-line input with line editing and command history, similar to readline. It uses raw terminal mode via `/dev/tty` and VT100 escape sequences. Unix/Linux only -- Windows is not supported.

## Build & Test Commands

```bash
go build ./...        # Build all packages
go test ./...         # Run all tests
go test -run TestName # Run a single test (gunit fixture name)
go vet ./...          # Static analysis
```

## Architecture

Single-package library (`package cle`) with three source files:

- **cle.go** - Core implementation: `CLE` struct, `ReadInput()` main loop, input handlers (arrow keys, control keys, delete, paste), history management, terminal I/O via `/dev/tty`
- **options.go** - Functional options pattern: `Option` type and configuration functions (`HistoryFile`, `HistorySize`, `TestMode`, etc.)
- **cle_test.go** - Tests using `github.com/smarty/gunit` framework
- **cmd/main.go** - Example CLI application

### Key Design Patterns

- **Input loop**: `ReadInput()` reads 3-byte chunks from the terminal and dispatches through a chain of handler methods (`handleArrowKeys` -> `handleDeleteKey` -> `handleControlKeys` -> `handleEnterKey` -> `handleAnySingleKey` -> `handlePaste`). Each handler returns `bool` to indicate if it consumed the input.
- **Functional options**: `NewCLE(options ...Option)` accepts variadic config functions.
- **TestMode**: Pass `TestMode(true)` to suppress terminal output, enabling unit tests to exercise input handling logic without a real TTY.
- **Receiver style**: Methods use `this` as the receiver name throughout.

### Dependencies

- `github.com/pkg/term` - Raw terminal mode and TTY operations
- `github.com/smarty/gunit` + `github.com/smarty/assertions` - Test framework (test only)
