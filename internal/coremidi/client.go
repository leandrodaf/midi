//go:build darwin
// +build darwin

package coremidi

/*
#cgo LDFLAGS: -framework CoreMIDI -framework CoreFoundation -framework CoreServices
#include <CoreMIDI/CoreMIDI.h>
#include <CoreFoundation/CFRunLoop.h>

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

// runMIDIRunLoop drives the current thread's CFRunLoop until CFRunLoopStop is
// called on it.  A 10-second iteration timeout is used so that a stale
// kCFRunLoopRunTimedOut result simply re-enters the loop rather than exiting,
// while kCFRunLoopRunStopped causes an immediate clean return.
static void runMIDIRunLoop(void) {
    for (;;) {
        int r = (int)CFRunLoopRunInMode(kCFRunLoopDefaultMode, 10.0, false);
        if (r == kCFRunLoopRunStopped) return;
    }
}
*/
import "C"
import (
	"fmt"
	"runtime"
	"runtime/cgo"
	"unsafe"
)

// Client wraps a CoreMIDI client reference and an optional notification channel.
type Client struct {
	client    C.MIDIClientRef
	NotifyCh  <-chan int32 // receives MIDINotificationMessageID values; nil if unsupported
	notifyCh  chan int32
	notifyHdl cgo.Handle
	runLoop   C.CFRunLoopRef // the CFRunLoop on which the MIDI client was created
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

// NewClient creates a CoreMIDI client on a dedicated OS thread that runs a
// CFRunLoop.
//
// CoreMIDI delivers setup-change notifications (device connect/disconnect) on
// the CFRunLoop of the thread that called MIDIClientCreate.  Go goroutines do
// not have a CFRunLoop by default, so without this fix the notification
// callback is registered but never invoked.
//
// The fix: lock a goroutine to a fresh OS thread, create the CoreMIDI client
// there, capture the thread's CFRunLoopRef, then block the thread on
// runMIDIRunLoop() for the lifetime of the client.  Dispose() stops the run
// loop, which unblocks the thread and allows it to exit cleanly.
func NewClient(name string) (Client, error) {
	type result struct {
		client Client
		err    error
	}
	resultCh := make(chan result, 1)

	go func() {
		// Bind this goroutine to a dedicated OS thread so that
		// CFRunLoopGetCurrent() returns a stable, thread-local run loop and
		// CoreMIDI can deliver notifications to it.
		// We intentionally do NOT defer runtime.UnlockOSThread(): the goroutine
		// stays alive driving the CFRunLoop, and Go terminates the locked OS
		// thread automatically when the goroutine returns.
		runtime.LockOSThread()

		notifyCh := make(chan int32, 16)
		h := cgo.NewHandle(notifyCh)

		var clientRef C.MIDIClientRef
		var createErr error

		stringToCFString(name, func(cfName C.CFStringRef) {
			osStatus := C.newMIDIClientWithNotify(cfName, unsafe.Pointer(uintptr(h)), &clientRef)
			if osStatus != C.noErr {
				createErr = fmt.Errorf("%d: failed to create a client", int(osStatus))
			}
		})

		if createErr != nil {
			h.Delete()
			resultCh <- result{err: createErr}
			return
		}

		rl := C.CFRunLoopGetCurrent()
		resultCh <- result{client: Client{
			client:    clientRef,
			NotifyCh:  notifyCh,
			notifyCh:  notifyCh,
			notifyHdl: h,
			runLoop:   rl,
		}}

		// Block this OS thread on the run loop so CoreMIDI can deliver
		// notifications.  Returns only when Dispose() calls CFRunLoopStop.
		C.runMIDIRunLoop()
	}()

	r := <-resultCh
	return r.client, r.err
}

// Dispose stops the CoreMIDI notification run loop and releases the cgo.Handle.
// After Dispose the Client must not be used again.
func (c *Client) Dispose() {
	// Stop the run loop first so the OS thread can exit cleanly before we
	// delete the handle it references.
	if c.runLoop != 0 {
		C.CFRunLoopStop(c.runLoop)
		c.runLoop = 0
	}
	if c.notifyHdl != 0 {
		c.notifyHdl.Delete()
		c.notifyHdl = 0
	}
}
