//go:build darwin
// +build darwin

package coremidi

/*
#cgo LDFLAGS: -framework CoreMIDI -framework CoreFoundation -framework CoreServices
#include <CoreMIDI/CoreMIDI.h>
#include <CoreServices/CoreServices.h>

// goMidiNotify is the exported Go callback; declared here so notifyBridgeFn can call it.
extern void goMidiNotify(int msgID, void *handle);

// notifyBridgeFn is the C-side bridge registered with MIDIClientCreate.
static void notifyBridgeFn(const MIDINotification *msg, void *refCon) {
    goMidiNotify((int)msg->messageID, refCon);
}

// newMIDIClientWithNotify wraps MIDIClientCreate with our bridge callback.
static OSStatus newMIDIClientWithNotify(CFStringRef name, void *refCon, MIDIClientRef *outClient) {
    return MIDIClientCreate(name, notifyBridgeFn, refCon, outClient);
}
*/
import "C"
import (
	"fmt"
	"runtime/cgo"
	"unsafe"
)

// Client wraps a CoreMIDI client reference and an optional notification channel.
type Client struct {
	client    C.MIDIClientRef
	NotifyCh  <-chan int32 // receives MIDINotificationMessageID values; nil if unsupported
	notifyCh  chan int32
	notifyHdl cgo.Handle
}

// goMidiNotify is called from the C notifyBridgeFn when CoreMIDI fires a
// setup-change notification. It forwards the message ID to the Go channel
// stored in the cgo.Handle.
//
//export goMidiNotify
func goMidiNotify(msgID C.int, handle unsafe.Pointer) {
	if handle == nil {
		return
	}
	h := cgo.Handle(uintptr(handle))
	ch, _ := h.Value().(chan int32)
	if ch == nil {
		return
	}
	select {
	case ch <- int32(msgID):
	default:
		// drop if consumer is slow — notification channel is best-effort
	}
}

// NewClient creates a CoreMIDI client with the given display name and registers
// a notification callback so that device-change events are forwarded to the
// returned Client.NotifyCh channel. The cgo.Handle embedded in Client must be
// released by calling Dispose() when the client is no longer needed.
func NewClient(name string) (client Client, err error) {
	notifyCh := make(chan int32, 16)
	h := cgo.NewHandle(notifyCh)

	stringToCFString(name, func(cfName C.CFStringRef) {
		var clientRef C.MIDIClientRef
		osStatus := C.newMIDIClientWithNotify(cfName, h.Pointer(), &clientRef)
		if osStatus != C.noErr {
			err = fmt.Errorf("%d: failed to create a client", int(osStatus))
		} else {
			client = Client{
				client:    clientRef,
				NotifyCh:  notifyCh,
				notifyCh:  notifyCh,
				notifyHdl: h,
			}
		}
	})

	if err != nil {
		h.Delete()
	}
	return
}

// Dispose releases the cgo.Handle associated with the notification channel.
// It must be called when the client is no longer needed to avoid a memory leak.
func (c *Client) Dispose() {
	if c.notifyHdl != 0 {
		c.notifyHdl.Delete()
		c.notifyHdl = 0
	}
}
