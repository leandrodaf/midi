# MIDI Client Library

A native Go library for capturing and manipulating MIDI events, supporting macOS and Windows operating systems without the need for external libraries or DLLs.

## Table of Contents

- [Introduction](#introduction)
- [Features](#features)
- [Installation](#installation)
- [Quick Usage](#quick-usage)
- [Configuration](#configuration)
- [Contribution](#contribution)
- [License](#license)

## Introduction

This project provides a **fully native** interface for working with MIDI devices, enabling the capture of events and filtering of MIDI commands without relying on any external libraries or dependencies. The library is designed to be easy to use and extensible, making it a straightforward choice for applications that require MIDI manipulation.

## Features

- **Native Support**: Works seamlessly on macOS and Windows without the need for additional libraries or DLLs.
- **Device Listing**: Easily list available MIDI devices connected to your system.
- **Device Selection**: Select MIDI devices for capturing events with simple function calls.
- **Event Capturing**: Capture MIDI events with support for filtering commands, allowing you to focus on the events that matter.
- **Built-in Logging**: Implemented logging for monitoring and debugging, providing insights into the MIDI event flow.

## Installation

To install the library, you can use the following Go command:

```bash
go get github.com/leandrodaf/midi
```

## Quick Usage

Here is a simple example of how to use the library to capture MIDI events:

```go
package main

import (
	"fmt"

	"github.com/leandrodaf/midi/internal/logger"
	"github.com/leandrodaf/midi/sdk/contracts"
	"github.com/leandrodaf/midi/sdk/midi"
)

func main() {
	log := logger.NewStandardLogger()

	client, err := midi.NewMIDIClient(
		contracts.WithLogger(log),
		contracts.WithLogLevel(contracts.InfoLevel),
		contracts.WithMIDIEventFilter(contracts.MIDIEventFilter{
			Commands: []contracts.MIDICommand{contracts.NoteOn, contracts.NoteOff},
		}),
	)
	if err != nil {
		log.Error("Failed to initialize MIDI client", log.Field().Error("error", err))
		return
	}

	devices, err := client.ListDevices()
	if err != nil || len(devices) == 0 {
		log.Error("No MIDI devices found or error listing devices", log.Field().Error("error", err))
		return
	}
	fmt.Println("Available MIDI devices:", devices)

	if err = client.SelectDevice(0); err != nil {
		log.Error("Failed to select MIDI device", log.Field().Error("error", err))
		return
	}

	eventChannel := make(chan contracts.MIDI, 100)
	go func() {
		for event := range eventChannel {
			log.Info("MIDI Event",
				log.Field().Uint64("Timestamp", event.Timestamp),
				log.Field().Int("Command", int(event.Command)),
				log.Field().Int("Note", int(event.Note)),
				log.Field().Int("Velocity", int(event.Velocity)),
			)
		}
	}()

	client.StartCapture(eventChannel)
	defer client.Stop()

	fmt.Println("Capturing MIDI events... Press Ctrl+C to exit.")
	select {} // Run indefinitely
}
```

## Configuration

The library allows for various configuration options when creating a MIDI client. Here are some of the available options:

- **Logger**: A custom logger can be provided.
- **LogLevel**: Logging level (Info, Debug, Error, etc.).
- **MIDIEventFilter**: A filter to specify which MIDI commands to capture.

Example configuration:

```go
client, err := midi.NewMIDIClient(
	contracts.WithLogger(log),
	contracts.WithLogLevel(contracts.InfoLevel),
	contracts.WithMIDIEventFilter(contracts.MIDIEventFilter{
		Commands: []contracts.MIDICommand{contracts.NoteOn, contracts.NoteOff},
	}),
)
```

## Contribution

Contributions are welcome! To contribute to the project, please follow these steps:

1. **Fork the repository**.
2. **Create a new branch** for your feature or fix:
   ```bash
   git checkout -b feature-your-feature-name
   ```
3. **Make your changes** and commit:
   ```bash
   git commit -m "Adds new feature"
   ```
4. **Push your changes to the remote repository**:
   ```bash
   git push origin feature-your-feature-name
   ```
5. **Create a Pull Request**.

## License

This project is licensed under the [MIT License](LICENSE).
