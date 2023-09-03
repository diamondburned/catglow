package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"machine"

	"tinygo.org/x/drivers/ws2812"
)

// Reader combines the io.ByteReader and io.Reader interfaces.
type Reader interface {
	io.ByteReader
	io.Reader
}

var _ Reader = (*machine.UART)(nil)

// Device stores the current state of the device.
type Device struct {
	uart *machine.UART
	led  ws2812.Device

	numLEDs uint16
}

// NewDevice creates a new device.
func NewDevice(uart *machine.UART, ledPin machine.Pin) *Device {
	return &Device{
		uart: uart,
		led:  ws2812.New(ledPin),
	}
}

// Run runs the device loop forever.
func (d *Device) Run() {
	for {
		p, err := d.readPacket()
		if err != nil {
			d.panic(err)
		}
		if err := d.handlePacket(p); err != nil {
			d.logError(err)
		}
	}
}

func (d *Device) panic(err error) {
	d.logError(err)
	d.sendPacket(PanicPacket{})
	panic("device panic")
}

func (d *Device) logError(err error) {
	d.sendPacket(ErrorPacket{Message: err.Error()})
}

func (d *Device) sendPacket(p OutgoingPacket) {
	switch p := p.(type) {
	case ErrorPacket:
		d.uart.WriteByte(byte(TypeErrorPacket))
		binary.Write(d.uart, binary.LittleEndian, uint16(len(p.Message)))
		io.WriteString(d.uart, p.Message)

	case PanicPacket:
		d.uart.WriteByte(byte(TypePanicPacket))
	}
}

func (d *Device) readPacket() (IncomingPacket, error) {
	ptype, err := d.uart.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read incoming packet type: %w", err)
	}

	switch IncomingPacketType(ptype) {
	case TypeInitializePacket:
		var p InitializePacket
		if err := binary.Read(d.uart, binary.LittleEndian, &p); err != nil {
			return nil, fmt.Errorf("failed to read number of LEDs: %w", err)
		}
		return p, nil

	case TypeClearPacket:
		var p ClearPacket
		return p, nil

	case TypeSetPacket:
		var p SetPacket
		p.Pix = make([]uint8, 3*d.numLEDs)
		if _, err := io.ReadFull(d.uart, p.Pix); err != nil {
			return nil, fmt.Errorf("failed to read pixel data: %w", err)
		}
		return p, nil

	default:
		return nil, fmt.Errorf("unknown packet type: %d", ptype)
	}
}

func (d *Device) handlePacket(p IncomingPacket) error {
	switch p := p.(type) {
	case InitializePacket:
		if p.NumLEDs < 1 {
			return fmt.Errorf("invalid number of LEDs: %d", p.NumLEDs)
		}
		d.numLEDs = p.NumLEDs
		return nil

	case ClearPacket:
		for i := 0; i < int(d.numLEDs); i++ {
			d.uart.WriteByte(0)
			d.uart.WriteByte(0)
			d.uart.WriteByte(0)
		}
		return nil

	case SetPacket:
		if len(p.Pix) != 3*int(d.numLEDs) {
			return fmt.Errorf("invalid number of pixels: %d", len(p.Pix)/3)
		}
		for _, b := range p.Pix {
			d.led.WriteByte(b)
		}
		return nil

	default:
		return fmt.Errorf("unknown packet type: %T", p)
	}
}
