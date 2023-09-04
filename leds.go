package catglow

import (
	"io"
	"unsafe"
)

// LEDs describes a strip of LEDs. It is a preallocated slice of RGBColor.
type LEDs []RGBColor

// NewLEDs creates a new strip of LEDs. Colors are initialized to black
// (off).
func NewLEDs(numLEDs int) LEDs {
	return make(LEDs, numLEDs)
}

// WriteTo implements io.WriterTo. It writes the LED strip to the given writer
// as a series of RGBColor values.
func (l LEDs) WriteTo(w io.Writer) (int64, error) {
	var written int64
	for _, c := range l {
		n, err := w.Write(c[:])
		written += int64(n)
		if err != nil {
			return written, err
		}
	}
	return written, nil
}

// AsPixels returns the LED strip as a slice of uint8 values. Each LED is
// represented by three values, one for each color channel.
func (l LEDs) AsPixels() []uint8 {
	return unsafe.Slice((*uint8)(unsafe.Pointer(&l[0])), 3*len(l))
}

// Set sets the color of the LED at the given index.
func (l LEDs) Set(i int, c RGBColor) {
	l[i] = c
}

// SetRange sets the color of the LEDs in the given range.
func (l LEDs) SetRange(start, end int, c RGBColor) {
	for i := start; i < end; i++ {
		l[i] = c
	}
}

// Draw draws the given LEDs into the strip at the given index.
// It stops when either l or other is exhausted and returns the number of LEDs
// written.
func (l LEDs) Draw(start int, other LEDs) int {
	for i := range other {
		if start+i >= len(l) {
			return i
		}
		l[start+i] = other[i]
	}
	return len(other)
}
