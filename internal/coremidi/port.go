//go:build darwin
// +build darwin

package coremidi

import (
	"fmt"
	"syscall"
	"unsafe"
)

/*
#cgo LDFLAGS: -framework CoreMIDI -framework CoreFoundation
#include <CoreMIDI/CoreMIDI.h>
#include <stdio.h>
#include <unistd.h>

static void MIDIInputProc(const MIDIPacketList *pktlist, void *readProcRefCon, void *srcConnRefCon)
{
  MIDIPacket *packet = (MIDIPacket *)&(pktlist->packet[0]);
  UInt32 packetCount = pktlist->numPackets;
  int i, n;
  Byte *data;

  int lengthBytes = 2;
  int timeStampBytes = 8;

  for (i = 0; i < packetCount; i++) {
    data = calloc(sizeof(Byte), packet->length + lengthBytes + timeStampBytes);

    memcpy(data, &(packet->length), lengthBytes);
    memcpy(data + lengthBytes, &(packet->timeStamp), timeStampBytes);
    memcpy(data + lengthBytes + timeStampBytes, packet->data, packet->length);

    n = write(*(int *)srcConnRefCon, data, packet->length + lengthBytes + timeStampBytes);
    packet = MIDIPacketNext(packet);
    free(data);
  }
}

typedef void (*midi_input_proc)(const MIDIPacketList *pktlist, void *readProcRefCon, void *srcConnRefCon);

static midi_input_proc getProc()
{
  return *MIDIInputProc;
}

*/
import "C"

type OutputPort struct {
	port C.MIDIPortRef
}

func NewOutputPort(client Client, name string) (outputPort OutputPort, err error) {
	var port C.MIDIPortRef

	stringToCFString(name, func(cfName C.CFStringRef) {
		osStatus := C.MIDIOutputPortCreate(client.client, cfName, &port)

		if osStatus != C.noErr {
			err = fmt.Errorf("%d: failed to create a port", int(osStatus))
		} else {
			outputPort = OutputPort{port}
		}
	})

	return
}

type ReadProc func(source Source, packet Packet)

type InputPort struct {
	port     C.MIDIPortRef
	readProc ReadProc
	writeFds []*C.int
}

func NewInputPort(client Client, name string, readProc ReadProc) (inputPort InputPort, err error) {
	var port C.MIDIPortRef

	stringToCFString(name, func(cfName C.CFStringRef) {
		osStatus := C.MIDIInputPortCreate(client.client,
			cfName,
			(C.MIDIReadProc)(C.getProc()),
			unsafe.Pointer(uintptr(0)),
			&port)

		if osStatus != C.noErr {
			err = fmt.Errorf("%d: failed to create a port", int(osStatus))
		} else {
			inputPort = InputPort{port, readProc, make([]*C.int, 0)}
		}
	})

	return
}

func (port InputPort) Connect(source Source) (PortConnection, error) {
	fd := make([]int, 2)
	syscall.Pipe(fd)
	readFd := fd[0]
	writeFd := C.int(fd[1])
	port.writeFds = append(port.writeFds, &writeFd)

	C.MIDIPortConnectSource(port.port, source.endpoint, unsafe.Pointer(&writeFd))

	go processIncomingPacket(
		readFd,
		func(data []byte, timeStamp uint64) {
			port.readProc(source, NewPacket(data, timeStamp))
		},
	)

	return PortConnection{port, source, &writeFd}, nil
}

type PortConnection struct {
	port    InputPort
	source  Source
	writeFd *C.int
}

func (connection PortConnection) Disconnect() {
	syscall.Close(int(*connection.writeFd))
	C.MIDIPortDisconnectSource(connection.port.port, connection.source.endpoint)
}
