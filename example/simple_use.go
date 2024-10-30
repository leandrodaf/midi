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
