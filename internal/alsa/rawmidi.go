//go:build linux && cgo
// +build linux,cgo

package alsa

/*
#cgo LDFLAGS: -lasound
#include <alsa/asoundlib.h>
#include <poll.h>
#include <stdlib.h>
#include <string.h>

#define MIDI_MAX_DEVICES 64

typedef struct {
	char name[256];
	char subdevice_name[256];
	char hw_addr[64];
} midi_input_info_t;

// enumerate_inputs fills `devices` with all available ALSA raw MIDI input
// subdevices found by scanning every sound card. Returns device count.
static int enumerate_inputs(midi_input_info_t *devices, int max) {
	int count = 0;
	int card  = -1;

	while (count < max && snd_card_next(&card) >= 0 && card >= 0) {
		snd_ctl_t *ctl = NULL;
		char hw[32];
		snprintf(hw, sizeof(hw), "hw:%d", card);
		if (snd_ctl_open(&ctl, hw, 0) < 0) continue;

		int dev = -1;
		for (;;) {
			if (snd_ctl_rawmidi_next_device(ctl, &dev) < 0 || dev < 0) break;

			snd_rawmidi_info_t *info = NULL;
			if (snd_rawmidi_info_malloc(&info) < 0) continue;

			snd_rawmidi_info_set_device(info, dev);
			snd_rawmidi_info_set_stream(info, SND_RAWMIDI_STREAM_INPUT);
			snd_rawmidi_info_set_subdevice(info, 0);

			if (snd_ctl_rawmidi_info(ctl, info) < 0) {
				snd_rawmidi_info_free(info);
				continue;
			}

			int nsubs = (int)snd_rawmidi_info_get_subdevices_count(info);
			if (nsubs < 1) nsubs = 1;

			for (int sub = 0; sub < nsubs && count < max; sub++) {
				if (sub > 0) {
					snd_rawmidi_info_set_subdevice(info, sub);
					if (snd_ctl_rawmidi_info(ctl, info) < 0) continue;
				}

				strncpy(devices[count].name,
				        snd_rawmidi_info_get_name(info), 255);
				devices[count].name[255] = '\0';

				const char *sn = snd_rawmidi_info_get_subdevice_name(info);
				if (sn && sn[0] != '\0')
					strncpy(devices[count].subdevice_name, sn, 255);
				else
					strncpy(devices[count].subdevice_name,
					        devices[count].name, 255);
				devices[count].subdevice_name[255] = '\0';

				snprintf(devices[count].hw_addr, 63,
				         "hw:%d,%d,%d", card, dev, sub);
				count++;
			}
			snd_rawmidi_info_free(info);
		}
		snd_ctl_close(ctl);
	}
	return count;
}

// rawmidi_fd returns the underlying poll file descriptor for an open handle.
static int rawmidi_fd(snd_rawmidi_t *handle) {
	struct pollfd pfd;
	if (snd_rawmidi_poll_descriptors(handle, &pfd, 1) < 1) return -1;
	return pfd.fd;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// DeviceInfo describes an ALSA raw MIDI input device.
type DeviceInfo struct {
	Name          string
	SubdeviceName string
	HWAddr        string // e.g. "hw:0,0,0"
}

// EnumerateInputs returns all available ALSA raw MIDI input devices.
func EnumerateInputs() ([]DeviceInfo, error) {
	var buf [C.MIDI_MAX_DEVICES]C.midi_input_info_t
	n := int(C.enumerate_inputs(&buf[0], C.int(C.MIDI_MAX_DEVICES)))
	if n < 0 {
		return nil, fmt.Errorf("ALSA enumeration failed (code %d)", n)
	}
	out := make([]DeviceInfo, n)
	for i := 0; i < n; i++ {
		out[i] = DeviceInfo{
			Name:          C.GoString(&buf[i].name[0]),
			SubdeviceName: C.GoString(&buf[i].subdevice_name[0]),
			HWAddr:        C.GoString(&buf[i].hw_addr[0]),
		}
	}
	return out, nil
}

// RawMIDI wraps an ALSA snd_rawmidi_t input handle.
type RawMIDI struct {
	handle *C.snd_rawmidi_t
}

// OpenInput opens an ALSA raw MIDI input device by hardware address (e.g. "hw:0,0,0").
// The handle is opened in blocking mode; Close() will unblock any pending Read().
func OpenInput(hwAddr string) (*RawMIDI, error) {
	cAddr := C.CString(hwAddr)
	defer C.free(unsafe.Pointer(cAddr))

	var h *C.snd_rawmidi_t
	ret := C.snd_rawmidi_open(&h, nil, cAddr, 0)
	if ret < 0 {
		return nil, fmt.Errorf("snd_rawmidi_open(%s): %s",
			hwAddr, C.GoString(C.snd_strerror(C.int(ret))))
	}
	return &RawMIDI{handle: h}, nil
}

// FD returns the underlying OS file descriptor, usable with unix.Poll/Select
// to implement cancellable reads without a tight loop.
func (r *RawMIDI) FD() (int, error) {
	fd := int(C.rawmidi_fd(r.handle))
	if fd < 0 {
		return -1, fmt.Errorf("could not obtain ALSA poll fd")
	}
	return fd, nil
}

// Read reads up to len(buf) raw MIDI bytes. Blocks until data arrives
// or the handle is closed (returns an error on close).
func (r *RawMIDI) Read(buf []byte) (int, error) {
	if len(buf) == 0 {
		return 0, nil
	}
	n := C.snd_rawmidi_read(r.handle, unsafe.Pointer(&buf[0]), C.size_t(len(buf)))
	if n < 0 {
		return 0, fmt.Errorf("snd_rawmidi_read: %s",
			C.GoString(C.snd_strerror(C.int(n))))
	}
	return int(n), nil
}

// Close releases the ALSA handle. Any goroutine blocked in Read will unblock
// with an error, allowing clean shutdown.
func (r *RawMIDI) Close() error {
	if r.handle == nil {
		return nil
	}
	ret := C.snd_rawmidi_close(r.handle)
	r.handle = nil
	if ret < 0 {
		return fmt.Errorf("snd_rawmidi_close: %s",
			C.GoString(C.snd_strerror(C.int(ret))))
	}
	return nil
}
