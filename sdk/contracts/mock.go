package contracts

import "context"

// MockMIDIClient is a configurable ClientMIDI mock for tests.
type MockMIDIClient struct {
	StartCaptureFunc func(ctx context.Context) (<-chan MIDI, error)
	StopFunc         func() error
	ListDevicesFunc  func() ([]DeviceInfo, error)
	SelectDeviceFunc func(deviceID int) error

	StartCaptureCalls int
	StopCalls         int
	ListDevicesCalls  int
	SelectDeviceCalls int
}

func (m *MockMIDIClient) StartCapture(ctx context.Context) (<-chan MIDI, error) {
	m.StartCaptureCalls++
	if m.StartCaptureFunc != nil {
		return m.StartCaptureFunc(ctx)
	}
	return nil, nil
}

func (m *MockMIDIClient) Stop() error {
	m.StopCalls++
	if m.StopFunc != nil {
		return m.StopFunc()
	}
	return nil
}

func (m *MockMIDIClient) ListDevices() ([]DeviceInfo, error) {
	m.ListDevicesCalls++
	if m.ListDevicesFunc != nil {
		return m.ListDevicesFunc()
	}
	return nil, nil
}

func (m *MockMIDIClient) SelectDevice(deviceID int) error {
	m.SelectDeviceCalls++
	if m.SelectDeviceFunc != nil {
		return m.SelectDeviceFunc(deviceID)
	}
	return nil
}
