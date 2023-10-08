package main

import (
	"io"
	"machine"
	"runtime"
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
	var n int
	for n < s.Buffered() && n < len(b) {
		c, err := s.ReadByte()
		if err != nil {
			return n, err
		}
		b[n] = c
		n++
	}

	// Emulate blocking-like behavior by yielding the scheduler.
	runtime.Gosched()
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
