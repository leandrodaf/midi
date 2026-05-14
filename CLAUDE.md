# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build ./...

# Run tests
go test ./...

# Run a single test
go test ./internal/midi/mididarwin/... -run TestName

# Lint (via pre-commit)
pre-commit run --all-files

# Run example
go run example/simple_use.go
```

Build tags are platform-specific: `darwin` for macOS (uses CoreMIDI via `go-coremidi`), `windows` for Windows (uses `winmm.dll` via syscalls). Dummy stubs in `*_dummy.go` files allow cross-compilation.

## Architecture

This is a Go library for capturing MIDI events natively on macOS and Windows without external DLLs.

### Layer structure

```
sdk/contracts/      → Public interfaces and types (ClientMIDI, Logger, Field, Option, MIDI, etc.)
sdk/midi/           → Public entry point: NewMIDIClient() applies options and dispatches to platform impl
internal/midi/      → Platform implementations (mididarwin, midiwindows)
internal/logger/    → ZapLogger: default Logger implementation backed by uber-zap
example/            → Usage example
```

### Key design decisions

**Platform dispatch** — `sdk/midi/midi_client_factory.go` maps `runtime.GOOS` to an initializer function. Adding a new platform means registering a new entry in `clientInitializers`.

**Options pattern** — `contracts.Option` is a `func(*ClientOptions)`. All user-facing config (logger, log level, log destination, channel buffer size, event filter, CoreMIDI client name) flows through functional options passed to `NewMIDIClient`.

**Event delivery** — `StartCapture(context.Context)` creates and returns a receive-only channel owned by the client. The channel is stored via `atomic.Value` for goroutine-safe access in the MIDI callback, closed on context cancellation or `Stop()`, and sized by `WithChannelBufferSize`. When the buffer is full, events are silently dropped with a warning log.

**Graceful shutdown** — The Darwin client uses `sync.WaitGroup` + `sync.Once` to ensure in-flight callbacks complete before `Stop()` returns. The Windows client uses `midiInStop`/`midiInClose` syscalls.

**Logger abstraction** — `contracts.Logger` is an interface and `contracts.Field` is a plain struct with standalone constructors like `contracts.StringField` and `contracts.ErrField`. The default logger writes plain-string output with JSON-encoded fields appended.

### MIDI event filtering

`MIDIEventFilter.Commands` is an allowlist of `MIDICommand` values (`NoteOn = 0x90`, `NoteOff = 0x80`). If `MIDIEventFilter` is nil, all commands pass through. The Windows implementation also strips the channel nibble (`status & 0xF0`) before filtering.
