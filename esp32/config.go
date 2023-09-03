package esp32

var (
	NumLEDs  = 192
	BackLEDs = [2]int{40, NumLEDs}
	SideLEDs = [2]int{0, BackLEDs[0]}
)

// EachLED calls f for each LED in the range [leds[0], leds[1]).
func EachLED(leds [2]int, f func(int)) {
	for i := leds[0]; i < leds[1]; i++ {
		f(i)
	}
}
