package contracts

// DeviceInfo contains information about a MIDI device.
type DeviceInfo struct {
	Name         string // Device name.
	Manufacturer string // Device manufacturer.
	EntityName   string // Name of the entity to which the device belongs.
}
