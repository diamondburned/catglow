package main

import (
	"image/color"
	"machine"
	"math"
	"runtime/interrupt"
	"time"

	"libdb.so/catglow/esp32"
	"tinygo.org/x/drivers/ws2812"
)

func main() {
	machine.GPIO27.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.LED.Configure(machine.PinConfig{Mode: machine.PinOutput})

	ledStrip := ws2812.New(machine.GPIO27)

	fullBright := make([]color.RGBA, esp32.NumLEDs)
	drawTransFlag(fullBright[esp32.BackLEDs[0]:esp32.BackLEDs[1]])

	currentBright := make([]color.RGBA, esp32.NumLEDs)
	copy(currentBright, fullBright)

	ticker := time.NewTicker(time.Second / 60)
	defer ticker.Stop()

	nextIntensity := newBreathingAnimation(breathingSine, time.Now(), time.Second*5)
	var statusLED bool

	for t := range ticker.C {
		statusLED = !statusLED
		machine.LED.Set(statusLED)

		intensity := uint(nextIntensity(t) * 0xFF)
		for i := range currentBright {
			currentBright[i].R = uint8(uint(fullBright[i].R) * intensity / 0xFF)
			currentBright[i].G = uint8(uint(fullBright[i].G) * intensity / 0xFF)
			currentBright[i].B = uint8(uint(fullBright[i].B) * intensity / 0xFF)
		}
		critical(func() { ledStrip.WriteColors(currentBright) })
	}
}

// drawTransFlag draws the transgender flag onto the given LEDs.
func drawTransFlag(leds []color.RGBA) {
	const chunks = 5

	chunkSize := len(leds) / chunks // 5 parts of the flag
	drawChunk := func(n int, c color.RGBA) {
		if n > chunks {
			panic("too many chunks")
		}
		start := n * chunkSize
		end := start + chunkSize
		for i := start; i < end; i++ {
			leds[i] = c
		}
	}

	drawChunk(0, color.RGBA{10, 150, 204, 255})
	drawChunk(1, color.RGBA{255, 94, 155, 255})
	drawChunk(2, color.RGBA{255, 255, 255, 255})
	drawChunk(3, color.RGBA{255, 94, 155, 255})
	drawChunk(4, color.RGBA{10, 150, 204, 255})
}

type breathingFunction uint8

const (
	breathingLinear breathingFunction = iota
	breathingSine
)

// newBreathingAnimation returns a function that returns the brightness of the
// LED at the given time. The returned function is meant to be called once per
// frame. The returned value is in the range [0, 1].
func newBreathingAnimation(function breathingFunction, start time.Time, duration time.Duration) func(time.Time) float64 {
	halfDuration := duration / 2
	halfDurationf := float64(halfDuration)

	switch function {
	case breathingLinear:
		return func(now time.Time) float64 {
			elapsed := now.Sub(start) % duration
			return math.Abs(1 - (float64(elapsed) / halfDurationf))
		}
	case breathingSine:
		return func(now time.Time) float64 {
			elapsed := now.Sub(start)
			return (1 + (math.Cos(float64(elapsed) / halfDurationf * math.Pi))) / 2
		}
	default:
		panic("unknown breathing function")
	}
}

func critical(f func()) {
	state := interrupt.Disable()
	f()
	interrupt.Restore(state)
}
