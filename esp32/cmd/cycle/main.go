package main

import (
	"image/color"
	"machine"
	"runtime/interrupt"
	"time"

	"libdb.so/catglow/esp32"
	"tinygo.org/x/drivers/ws2812"
)

var leds = make([]color.RGBA, esp32.NumLEDs)
var cycles = []color.RGBA{
	{10, 150, 204, 255},
	{255, 255, 255, 255},
	{255, 94, 155, 255},
}

func main() {
	machine.GPIO27.Configure(machine.PinConfig{Mode: machine.PinOutput})

	led := ws2812.New(machine.GPIO27)
	var cycle int

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		current := cycles[cycle]
		cycle = (cycle + 1) % len(cycles)

		esp32.EachLED(esp32.BackLEDs, func(i int) { leds[i] = current })
		esp32.EachLED(esp32.SideLEDs, func(i int) { leds[i] = color.RGBA{} })

		critical(func() { led.WriteColors(leds[:]) })
	}
}

func critical(f func()) {
	state := interrupt.Disable()
	f()
	interrupt.Restore(state)
}
