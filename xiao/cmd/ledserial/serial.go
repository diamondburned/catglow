package main

import (
	"io"
	"machine"
	"runtime"
	"time"
)

type serialIO struct {
	machine.Serialer
}

// ByteReadWriter describes a device that can read and write bytes.
// Usually, machine.Serialer implements this interface.
type ByteReadWriter interface {
	ReadByte() (byte, error)
	WriteByte(byte) error
}

// SerialReadWriter combines the io.ReadWriter and machine.Serialer interfaces.
type SerialReadWriter interface {
	io.ReadWriter
	ByteReadWriter
	// Buffered returns the number of bytes currently buffered in the serial
	// device.
	Buffered() int
}

var _ ByteReadWriter = machine.Serialer(nil)

// WrapSerial wraps a machine.Serialer in an io.ReadWriter.
func WrapSerial(serial machine.Serialer) SerialReadWriter {
	return serialIO{Serialer: serial}
}

func (s serialIO) Read(b []byte) (int, error) {
	n := s.Buffered()
	if n > 0 {
		for i := 0; i < n; i++ {
			c, err := s.ReadByte()
			if err != nil {
				return i, err
			}
			b[i] = c
		}
		// Emulate blocking-like behavior by yielding the scheduler.
		runtime.Gosched()
	} else {
		// Sleep to reduce CPU usage.
		time.Sleep(1 * time.Millisecond)
	}
	return n, nil
}

func (s serialIO) Write(b []byte) (int, error) {
	for _, c := range b {
		if err := s.WriteByte(c); err != nil {
			return 0, err
		}
	}
	runtime.Gosched()
	return len(b), nil
}
