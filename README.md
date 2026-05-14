# MIDI Client Library

A native Go library for capturing and manipulating MIDI events on macOS and Windows without external DLLs.

## Table of Contents

- [Introduction](#introduction)
- [Features](#features)
- [Installation](#installation)
- [Quick Usage](#quick-usage)
- [Configuration](#configuration)
- [Contribution](#contribution)
- [License](#license)

## Introduction

This project provides a fully native interface for working with MIDI devices, enabling event capture and MIDI command filtering without external dependencies.

## Features

- Native support for macOS and Windows.
- List available MIDI input devices.
- Select a device and capture MIDI events.
- Filter incoming MIDI commands.
- Structured logging with configurable level, destination, and channel buffer size.

## Installation

```bash
go get github.com/leandrodaf/midi
```

## Quick Usage

```go
package main

import (
    "context"
    "fmt"
    "os/signal"
    "syscall"

    "github.com/leandrodaf/midi/internal/logger"
    "github.com/leandrodaf/midi/sdk/contracts"
    "github.com/leandrodaf/midi/sdk/midi"
)

func main() {
    log := logger.NewLogger()

    client, err := midi.NewMIDIClient(
        contracts.WithLogger(log),
        contracts.WithLogLevel(contracts.InfoLevel),
        contracts.WithChannelBufferSize(100),
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
        log.Info("MIDI Event",
            contracts.Uint64Field("Timestamp", event.Timestamp),
            contracts.IntField("Command", int(event.Command)),
            contracts.IntField("Note", int(event.Note)),
            contracts.IntField("Velocity", int(event.Velocity)),
        )
    }
}
```

## Configuration

Available options include:

- `WithLogger` to inject a custom logger.
- `WithLogLevel` to control the minimum log level.
- `WithLogDestination` to write logs to console or file.
- `WithChannelBufferSize` to size the event channel returned by `StartCapture`.
- `WithMIDIEventFilter` to allow only selected MIDI commands.
- `WithCoreMIDIConfig` to customize the CoreMIDI client name on macOS.

```go
client, err := midi.NewMIDIClient(
    contracts.WithLogger(log),
    contracts.WithLogLevel(contracts.InfoLevel),
    contracts.WithLogDestination(contracts.ConsoleLog),
    contracts.WithChannelBufferSize(256),
    contracts.WithMIDIEventFilter(contracts.MIDIEventFilter{
        Commands: []contracts.MIDICommand{contracts.NoteOn, contracts.NoteOff},
    }),
)
```

## Contribution

Contributions are welcome. Fork the repository, create a branch, make your changes, and open a pull request.

## License

This project is licensed under the [MIT License](LICENSE).
