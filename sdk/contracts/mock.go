package contracts

import "context"

// MockMIDIClient is a configurable ClientMIDI mock for tests.
// Set the *Func fields to override behaviour; the *Calls fields count invocations.
// Zero-value Func fields fall back to safe no-op defaults (WatchDevices closes
// the channel when ctx is cancelled; all others return nil/zero values).
type MockMIDIClient struct {
	StartCaptureFunc func(ctx context.Context) (<-chan MIDI, error)
	StopFunc         func() error
	ListDevicesFunc  func() ([]DeviceInfo, error)
	SelectDeviceFunc func(deviceID int) error
	// WatchDevicesFunc is called by WatchDevices. When nil, the default
	// implementation returns a channel that is closed when ctx is cancelled.
	WatchDevicesFunc func(ctx context.Context) (<-chan DeviceEvent, error)

	StartCaptureCalls int
	StopCalls         int
	ListDevicesCalls  int
	SelectDeviceCalls int
	// WatchDevicesCalls is incremented on every call to WatchDevices.
	WatchDevicesCalls int
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

// WatchDevices delegates to WatchDevicesFunc when set; otherwise it returns a
// channel that is closed when ctx is cancelled, matching the contract that
// callers must range over the channel until it is closed.
func (m *MockMIDIClient) WatchDevices(ctx context.Context) (<-chan DeviceEvent, error) {
	m.WatchDevicesCalls++
	if m.WatchDevicesFunc != nil {
		return m.WatchDevicesFunc(ctx)
	}
	ch := make(chan DeviceEvent)
	go func() {
		<-ctx.Done()
		close(ch)
	}()
	return ch, nil
}
