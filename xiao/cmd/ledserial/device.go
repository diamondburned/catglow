package main

import (
	"fmt"
	"machine"

	"libdb.so/catglow/ledserial"
	"tinygo.org/x/drivers/ws2812"
)

// Device stores the current state of the device.
type Device struct {
	serial SerialReadWriter
	led    ws2812.Device

	ledBuffer []byte
}

// NewDevice creates a new device.
func NewDevice(serial machine.Serialer, ledPin machine.Pin) *Device {
	ledPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	return &Device{
		serial: WrapSerial(serial),
		led:    ws2812.New(ledPin),
	}
}

// Run runs the device loop forever.
func (d *Device) Run() {
	for {
		p, err := d.readPacket()
		if err != nil {
			d.logError(err)
			continue
		}

		d.log(fmt.Sprintf("received packet: %s", p.Type()))

		if err := d.handlePacket(p); err != nil {
			d.logError(err)
		}
	}
}

func (d *Device) log(msg string) {
	d.sendPacket(ledserial.LogPacket{Message: msg})
}

func (d *Device) logError(err error) {
	d.sendPacket(ledserial.ErrorPacket{Message: err.Error()})
}

func (d *Device) sendPacket(p ledserial.OutgoingPacket) {
	ledserial.WriteOutgoingPacket(d.serial, p)
}

func (d *Device) readPacket() (ledserial.IncomingPacket, error) {
	turnOnMainLED(255, 255, 255)

	p, err := ledserial.ReadIncomingPacket(d.serial, ledserial.ReadContext{
		LEDBuffer: d.ledBuffer,
	})

	turnOffMainLED()
	return p, err
}

func (d *Device) handlePacket(p ledserial.IncomingPacket) error {
	switch p := p.(type) {
	case ledserial.InitializePacket:
		if p.NumLEDs < 1 {
			return fmt.Errorf("invalid number of LEDs: %d", p.NumLEDs)
		}
		d.ledBuffer = make([]byte, 3*int(p.NumLEDs))
		d.clearLEDs(true)

	case ledserial.ClearPacket:
		d.clearLEDs(false)

	case ledserial.SetPacket:
		for _, b := range p.Pix {
			d.led.WriteByte(b)
		}

	default:
		return fmt.Errorf("unknown packet type: %T", p)
	}

	d.sendPacket(ledserial.AckPacket{
		IncomingPacketType: p.Type(),
	})
	return nil
}

func (d *Device) clearLEDs(signalReady bool) {
	var i int

	if signalReady {
		writeLEDRGB(d.led, 255, 0, 0) // red
		i++
	}

	for i < len(d.ledBuffer)-1 {
		writeLEDRGB(d.led, 0, 0, 0)
		i++
	}

	if signalReady {
		writeLEDRGB(d.led, 0, 0, 255) // blue
	} else {
		writeLEDRGB(d.led, 0, 0, 0)
	}
}

func writeLEDRGB(led ws2812.Device, r, g, b uint8) {
	led.WriteByte(r)
	led.WriteByte(g)
	led.WriteByte(b)
}
