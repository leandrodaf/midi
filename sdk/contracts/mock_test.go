package contracts_test

import (
	"context"
	"errors"
	"testing"

	"github.com/leandrodaf/midi/v2/sdk/contracts"
)

var _ contracts.ClientMIDI = (*contracts.MockMIDIClient)(nil)

func TestMockMIDIClient_UsesConfiguredFuncs(t *testing.T) {
	expectedChannel := make(chan contracts.MIDI)
	expectedStartErr := errors.New("start error")
	expectedStopErr := errors.New("stop error")
	expectedListErr := errors.New("list error")
	expectedSelectErr := errors.New("select error")
	expectedDevices := []contracts.DeviceInfo{{Name: "device-1"}}

	startCalled := false
	stopCalled := false
	listCalled := false
	selectCalled := false

	client := &contracts.MockMIDIClient{
		StartCaptureFunc: func(ctx context.Context) (<-chan contracts.MIDI, error) {
			startCalled = true
			return expectedChannel, expectedStartErr
		},
		StopFunc: func() error {
			stopCalled = true
			return expectedStopErr
		},
		ListDevicesFunc: func() ([]contracts.DeviceInfo, error) {
			listCalled = true
			return expectedDevices, expectedListErr
		},
		SelectDeviceFunc: func(deviceID int) error {
			selectCalled = true
			if deviceID != 7 {
				t.Fatalf("expected device ID 7, got %d", deviceID)
			}
			return expectedSelectErr
		},
	}

	channel, err := client.StartCapture(context.Background())
	if !startCalled {
		t.Fatalf("expected StartCaptureFunc to be called")
	}
	if channel != expectedChannel {
		t.Errorf("expected StartCapture to return configured channel")
	}
	if err != expectedStartErr {
		t.Errorf("expected StartCapture to return configured error")
	}

	err = client.Stop()
	if !stopCalled {
		t.Fatalf("expected StopFunc to be called")
	}
	if err != expectedStopErr {
		t.Errorf("expected Stop to return configured error")
	}

	devices, err := client.ListDevices()
	if !listCalled {
		t.Fatalf("expected ListDevicesFunc to be called")
	}
	if len(devices) != len(expectedDevices) || devices[0].Name != expectedDevices[0].Name {
		t.Errorf("expected ListDevices to return configured devices")
	}
	if err != expectedListErr {
		t.Errorf("expected ListDevices to return configured error")
	}

	err = client.SelectDevice(7)
	if !selectCalled {
		t.Fatalf("expected SelectDeviceFunc to be called")
	}
	if err != expectedSelectErr {
		t.Errorf("expected SelectDevice to return configured error")
	}
}

func TestMockMIDIClient_ZeroValuesWhenFuncsNil(t *testing.T) {
	client := &contracts.MockMIDIClient{}

	channel, err := client.StartCapture(context.Background())
	if channel != nil {
		t.Errorf("expected nil channel when StartCaptureFunc is not set")
	}
	if err != nil {
		t.Errorf("expected nil error when StartCaptureFunc is not set, got %v", err)
	}

	devices, err := client.ListDevices()
	if devices != nil {
		t.Errorf("expected nil devices when ListDevicesFunc is not set")
	}
	if err != nil {
		t.Errorf("expected nil error when ListDevicesFunc is not set, got %v", err)
	}

	err = client.SelectDevice(1)
	if err != nil {
		t.Errorf("expected nil error when SelectDeviceFunc is not set, got %v", err)
	}

	err = client.Stop()
	if err != nil {
		t.Errorf("expected nil error when StopFunc is not set, got %v", err)
	}
}

func TestMockMIDIClient_CallCountersIncrement(t *testing.T) {
	client := &contracts.MockMIDIClient{}

	_, _ = client.StartCapture(context.Background())
	_, _ = client.StartCapture(context.Background())
	_ = client.Stop()
	_, _ = client.ListDevices()
	_, _ = client.ListDevices()
	_ = client.SelectDevice(1)
	_ = client.SelectDevice(2)
	_ = client.SelectDevice(3)

	if client.StartCaptureCalls != 2 {
		t.Errorf("expected StartCaptureCalls to be 2, got %d", client.StartCaptureCalls)
	}
	if client.StopCalls != 1 {
		t.Errorf("expected StopCalls to be 1, got %d", client.StopCalls)
	}
	if client.ListDevicesCalls != 2 {
		t.Errorf("expected ListDevicesCalls to be 2, got %d", client.ListDevicesCalls)
	}
	if client.SelectDeviceCalls != 3 {
		t.Errorf("expected SelectDeviceCalls to be 3, got %d", client.SelectDeviceCalls)
	}
}
