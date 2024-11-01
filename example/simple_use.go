package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/leandrodaf/midi/internal/logger"
	"github.com/leandrodaf/midi/sdk/contracts"
	"github.com/leandrodaf/midi/sdk/midi"
)

func main() {
	log := logger.NewZapLogger()

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
	var wg sync.WaitGroup

	// Goroutine para processar eventos MIDI
	wg.Add(1)
	go func() {
		defer wg.Done()
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

	// Configurar canais para sinal de interrupção e conclusão
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan struct{}) // Canal para sinalizar que devemos encerrar
	closeOnce := sync.Once{}    // Garantir que o canal done seja fechado apenas uma vez

	// Função para encerrar a captura e sinalizar conclusão
	stopCapture := func(reason string) {
		log.Info(reason)
		client.Stop()
		closeOnce.Do(func() {
			close(eventChannel) // Fecha o canal de eventos para parar o goroutine de processamento
			close(done)         // Sinaliza que devemos encerrar
		})
	}

	// Goroutine para lidar com sinais de interrupção
	go func() {
		<-sigChan
		stopCapture("Received shutdown signal, stopping capture...")
	}()

	// Goroutine para lidar com o timeout
	go func() {
		time.Sleep(5 * time.Second) // Simula um período de captura curto
		stopCapture("Timeout reached, stopping capture...")
	}()

	fmt.Println("Capturing MIDI events... Press Ctrl+C to exit.")
	<-done // Aguarda até que o canal done seja fechado

	// Aguarda a conclusão do processamento de eventos
	wg.Wait()
	log.Info("Program terminated gracefully.")
}
