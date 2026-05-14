# MIDI Client Library

A native Go library for capturing MIDI events on macOS and Windows without external DLLs.

[![CI](https://github.com/leandrodaf/midi/actions/workflows/ci.yml/badge.svg)](https://github.com/leandrodaf/midi/actions/workflows/ci.yml)

## Table of Contents

- [Introduction](#introduction)
- [Features](#features)
- [Installation](#installation)
- [Quick Usage](#quick-usage)
- [Configuration](#configuration)
- [Filtering MIDI Commands](#filtering-midi-commands)
- [Custom Logger](#custom-logger)
- [Testing with MockMIDIClient](#testing-with-mockmidiClient)
- [Platform Notes](#platform-notes)
- [Contribution](#contribution)
- [License](#license)

## Introduction

This project provides a fully native interface for working with MIDI devices, enabling event capture and MIDI command filtering without external dependencies. macOS uses CoreMIDI via CGo; Windows uses `winmm.dll` via pure-Go syscalls.

## Features

- Native support for macOS (CoreMIDI) and Windows (winmm.dll).
- List available MIDI input devices.
- Select a device and capture MIDI events over a Go channel.
- Context-based lifecycle — cancel the context to stop capture and close the channel.
- Filter incoming MIDI commands via an allowlist.
- Structured logging with configurable level and destination.
- `MockMIDIClient` for easy unit testing in consumer code.

## Installation

```bash
go get github.com/leandrodaf/midi/v2@v2.0.2
```

## Quick Usage

```go
package main

import (
    "context"
    "fmt"
    "os/signal"
    "syscall"

    "github.com/leandrodaf/midi/v2/sdk/contracts"
    "github.com/leandrodaf/midi/v2/sdk/logger"
    "github.com/leandrodaf/midi/v2/sdk/midi"
)

func main() {
    log := logger.NewLogger()

    client, err := midi.NewMIDIClient(
        contracts.WithLogger(log),
        contracts.WithMIDIEventFilter(contracts.MIDIEventFilter{
            Commands: []contracts.MIDICommand{contracts.NoteOn, contracts.NoteOff},
        }),
    )
    if err != nil {
        log.Error("Failed to initialize MIDI client", contracts.ErrField("error", err))
        return
    }

    devices, err := client.ListDevices()
    if err != nil || len(devices) == 0 {
        log.Error("No MIDI devices found", contracts.ErrField("error", err))
        return
    }
    fmt.Println("Available MIDI devices:", devices)

    if err = client.SelectDevice(0); err != nil {
        log.Error("Failed to select MIDI device", contracts.ErrField("error", err))
        return
    }

    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    events, err := client.StartCapture(ctx)
    if err != nil {
        log.Error("Failed to start MIDI capture", contracts.ErrField("error", err))
        return
    }

    fmt.Println("Capturing MIDI events... Press Ctrl+C to exit.")
    for event := range events {
        fmt.Printf("cmd=0x%02X note=%d velocity=%d\n", event.Command, event.Note, event.Velocity)
    }
}
```

## Configuration

All options are passed to `midi.NewMIDIClient(opts...)`:

| Option | Default | Description |
|---|---|---|
| `WithLogger(l)` | built-in stderr logger | Inject a custom `contracts.Logger` |
| `WithLogLevel(level)` | `InfoLevel` | Minimum log level to emit |
| `WithLogDestination(dest, path...)` | `ConsoleLog` | `ConsoleLog` or `FileLog` (requires path) |
| `WithChannelBufferSize(n)` | `100` | Buffer size of the event channel |
| `WithMIDIEventFilter(f)` | nil (all commands) | Allowlist of MIDI commands to forward |
| `WithCoreMIDIConfig(c)` | client name `"GO MIDI Client"` | macOS-only CoreMIDI client name |

```go
client, err := midi.NewMIDIClient(
    contracts.WithLogLevel(contracts.DebugLevel),
    contracts.WithLogDestination(contracts.FileLog, "/var/log/midi.log"),
    contracts.WithChannelBufferSize(256),
)
```

**Log levels** (in increasing severity): `DebugLevel`, `InfoLevel`, `WarnLevel`, `ErrorLevel`, `FatalLevel`.

## Filtering MIDI Commands

Pass a `MIDIEventFilter` to receive only the commands you care about. Without a filter, all commands are forwarded.

```go
contracts.WithMIDIEventFilter(contracts.MIDIEventFilter{
    Commands: []contracts.MIDICommand{contracts.NoteOn, contracts.NoteOff},
})
```

## Custom Logger

Implement `contracts.Logger` to integrate with your own logging framework:

```go
type myLogger struct{}

func (l *myLogger) Info(msg string, fields ...contracts.Field)  { /* ... */ }
func (l *myLogger) Debug(msg string, fields ...contracts.Field) { /* ... */ }
func (l *myLogger) Warn(msg string, fields ...contracts.Field)  { /* ... */ }
func (l *myLogger) Error(msg string, fields ...contracts.Field) { /* ... */ }
func (l *myLogger) Fatal(msg string, fields ...contracts.Field) { /* ... */ }
func (l *myLogger) SetLevel(level contracts.LogLevel)           { /* ... */ }
func (l *myLogger) SetDestination(dest contracts.LogDestination, path ...string) { /* ... */ }

client, err := midi.NewMIDIClient(contracts.WithLogger(&myLogger{}))
```

Field constructors: `contracts.StringField`, `contracts.IntField`, `contracts.BoolField`,
`contracts.ErrField`, `contracts.Float64Field`, `contracts.TimeField`, `contracts.Uint64Field`, `contracts.Uint8Field`, `contracts.Int64Field`.

## Testing with MockMIDIClient

`contracts.MockMIDIClient` is provided for use in your own tests:

```go
mock := &contracts.MockMIDIClient{
    StartCaptureFunc: func(ctx context.Context) (<-chan contracts.MIDI, error) {
        ch := make(chan contracts.MIDI, 1)
        ch <- contracts.MIDI{Command: 0x90, Note: 60, Velocity: 100}
        close(ch)
        return ch, nil
    },
}

events, _ := mock.StartCapture(context.Background())
for e := range events {
    fmt.Println(e)
}
fmt.Println("StartCapture called:", mock.StartCaptureCalls) // 1
```

## Platform Notes

| Platform | Implementation | CGo |
|---|---|---|
| macOS | CoreMIDI via CGo (`-framework CoreMIDI`) | Required |
| Windows | `winmm.dll` via pure-Go syscalls | Not required |
| Linux / other | Stub — methods return errors | Not required |

The library compiles on all platforms. On unsupported platforms the client is a no-op stub that returns errors from every method.

## Contribution

Contributions are welcome. Fork the repository, create a branch, make your changes, and open a pull request.

## License

This project is licensed under the [MIT License](LICENSE).
