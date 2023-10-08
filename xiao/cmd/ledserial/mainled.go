package main

import (
	"machine"

	"tinygo.org/x/drivers/ws2812"
)

var mainLED ws2812.Device
var mainLEDPower = machine.GPIO11
var mainLEDInitialized bool

func initMainLED() {
	if !mainLEDInitialized {
		// https://wiki.seeedstudio.com/XIAO-RP2040-with-Arduino/
		mainLEDPower.Configure(machine.PinConfig{Mode: machine.PinOutput})
		mainLEDPower.Low()

		machine.GPIO12.Configure(machine.PinConfig{Mode: machine.PinOutput})
		mainLED = ws2812.New(machine.GPIO12)

		mainLEDInitialized = true
	}
}

func turnOnMainLED(r, g, b uint8) {
	initMainLED()
	mainLEDPower.High()
	mainLED.WriteByte(r)
	mainLED.WriteByte(g)
	mainLED.WriteByte(b)
}

func turnOffMainLED() {
	initMainLED()
	mainLEDPower.Low()
}
