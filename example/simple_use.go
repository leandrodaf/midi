package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/leandrodaf/midi/sdk/logger"
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

	log.Info("Capture ended.")
}
