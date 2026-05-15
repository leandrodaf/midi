//go:build linux && cgo
// +build linux,cgo

// Package midilinux — test-only exports for white-box testing of internal helpers.
package midilinux

import "github.com/leandrodaf/midi/v2/sdk/contracts"

// DiffDevicesExported is a test-only shim that calls the unexported diffDevices helper.
func DiffDevicesExported(prev, curr []contracts.DeviceInfo, evCh chan<- contracts.DeviceEvent) {
	diffDevices(prev, curr, evCh)
}
