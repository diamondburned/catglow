package main

import (
	"fmt"
	"io"
	"machine"
	"time"

	"libdb.so/catglow/ledserial"
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
	ledPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	return &Device{
		uart: uart,
		led:  ws2812.New(ledPin),
	}
}

// Run runs the device loop forever.
func (d *Device) Run() {
	go func() {
		for t := range time.Tick(5 * time.Second) {
			d.log(fmt.Sprintf("the time is now %s", t))
		}
	}()

	for {
		time.Sleep(1 * time.Second)

		p, err := d.readPacket()
		if err != nil {
			d.logError(err)
			continue
		}

		if err := d.handlePacket(p); err != nil {
			d.logError(err)
		}
	}
}

func (d *Device) panic(err error) {
	d.logError(err)
	d.sendPacket(ledserial.PanicPacket{})
	panic("device panic")
}

func (d *Device) log(msg string) {
	d.sendPacket(ledserial.LogPacket{Message: msg})
}

func (d *Device) logError(err error) {
	d.sendPacket(ledserial.ErrorPacket{Message: err.Error()})
}

func (d *Device) sendPacket(p ledserial.OutgoingPacket) {
	ledserial.WriteOutgoingPacket(d.uart, p)
}

func (d *Device) readPacket() (ledserial.IncomingPacket, error) {
	return ledserial.ReadIncomingPacket(d.uart, ledserial.ReadContext{
		NumLEDs: d.numLEDs,
	})
}

func (d *Device) handlePacket(p ledserial.IncomingPacket) error {
	switch p := p.(type) {
	case ledserial.InitializePacket:
		if p.NumLEDs < 1 {
			return fmt.Errorf("invalid number of LEDs: %d", p.NumLEDs)
		}
		d.numLEDs = p.NumLEDs
		d.clearLEDs()
		return nil

	case ledserial.ClearPacket:
		d.clearLEDs()
		return nil

	case ledserial.SetPacket:
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

func (d *Device) clearLEDs() {
	for i := 0; i < int(d.numLEDs); i++ {
		d.uart.WriteByte(0)
		d.uart.WriteByte(0)
		d.uart.WriteByte(0)
	}
}
