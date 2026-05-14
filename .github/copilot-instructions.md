# Copilot Instructions

## Commands

```bash
# Build
go build ./...

# Test all
go test ./...

# Run a single test (adjust package path and test name)
go test ./internal/midi/mididarwin/... -run TestName

# Lint (requires pre-commit)
pre-commit run --all-files

# Run example
go run example/simple_use.go
```

## Architecture

A native Go library for capturing MIDI events on macOS and Windows without external DLLs.

### Layer structure

```
sdk/contracts/      → Public interfaces and types (ClientMIDI, Logger, Field, Option, MIDI, etc.)
sdk/midi/           → Public entry point: NewMIDIClient() applies options, dispatches to platform impl
internal/midi/      → Platform implementations (mididarwin, midiwindows)
internal/coremidi/  → Low-level CoreMIDI bindings (Darwin only)
internal/logger/    → ZapLogger: default Logger implementation
example/            → Usage example
```

### Platform dispatch

`sdk/midi/midi_client_factory.go` maps `runtime.GOOS` to an initializer function via `clientInitializers`. To add a new platform, register a new `func(*contracts.ClientOptions) (contracts.ClientMIDI, error)` entry in that map.

Build tags control compilation: `//go:build darwin` / `//go:build windows`. Every platform package has a paired `*_dummy.go` (no build tag constraint) with stub implementations so the package compiles on other OSes.

### Key conventions

**Options pattern** — `contracts.Option` is `func(*ClientOptions)`. All configuration flows through functional options (`WithLogger`, `WithLogLevel`, `WithLogDestination`, `WithChannelBufferSize`, `WithMIDIEventFilter`, `WithCoreMIDIConfig`) passed to `NewMIDIClient`.

**Event delivery** — `StartCapture(context.Context)` returns a client-owned receive-only channel. The channel is stored in an `atomic.Value` for goroutine-safe access from MIDI callbacks, closed on context cancellation or `Stop()`, and sized with `WithChannelBufferSize`. When the buffer is full, events are silently dropped with a warning log.

**Graceful shutdown** — Darwin uses `sync.WaitGroup` + `sync.Once`; `Stop()` disconnects the port, swaps in a dummy channel to absorb any in-flight writes, then calls `wg.Wait()`. Windows uses `midiInStop`/`midiInClose` syscalls.

**Logger abstraction** — `contracts.Logger` is an interface and `contracts.Field` is a plain struct with standalone constructors such as `contracts.StringField` and `contracts.ErrField`. The default logger (`internal/logger`) emits plain-string output with JSON-encoded fields appended.

**MIDI event filtering** — `MIDIEventFilter.Commands` is an allowlist of `MIDICommand` bytes (`NoteOn = 0x90`, `NoteOff = 0x80`). A `nil` filter passes all commands via `contracts.IsCommandAllowed`. The Windows implementation strips the channel nibble (`status & 0xF0`) before comparing; Darwin compares the raw byte.
