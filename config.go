package catglow

import (
	"encoding"
	"fmt"
	"io"
	"time"

	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"
	"libdb.so/catglow/internal/led"
)

// Config is the configuration for the Catglow server.
type Config struct {
	// Device is the path to the device file for catglow.
	// This is usually /dev/ttyUSB0 or /dev/ttyACM0.
	Device string `toml:"device"`
	// Baud is the baud rate for the serial connection.
	Baud int `toml:"baud"`
	// Rate is the refresh rate for the LEDs.
	Rate int `toml:"rate"`
	// LEDs is a list of LED configurations.
	LEDs []LEDConfig `toml:"led"`
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.NumLEDs() == 0 {
		return errors.New("no LEDs configured")
	}

	// Check for overlapping LED ranges.
	for i, led1 := range c.LEDs {
		for j, led2 := range c.LEDs {
			if i == j {
				continue
			}

			if led1.Range[0] >= led2.Range[0] && led1.Range[0] <= led2.Range[1] {
				return fmt.Errorf("LED range %v overlaps with %v", led1.Range, led2.Range)
			}

			if led1.Range[1] >= led2.Range[0] && led1.Range[1] <= led2.Range[1] {
				return fmt.Errorf("LED range %v overlaps with %v", led1.Range, led2.Range)
			}
		}
	}

	return nil
}

// NumLEDs returns the number of LEDs configured.
func (c *Config) NumLEDs() int {
	var numLEDs int
	for _, led := range c.LEDs {
		if led.Range[1] > numLEDs {
			numLEDs = led.Range[1]
		}
	}
	return numLEDs
}

// LEDConfig is the configuration for a range of LEDs.
type LEDConfig struct {
	// Range is the range of LEDs to configure.
	Range [2]int `toml:"range"`

	// Only one of the following fields should be set.
	// If none are set, then the LEDs are unchanged.

	// Color is the color to set the LEDs to.
	Color *led.RGBColor `toml:"color,omitempty"`
	// Snake is the configuration for the snake animation.
	Snake *SnakeAnimationConfig `toml:"snake,omitempty"`
	// Visualizer is the configuration for the visualizer.
	Visualizer *VisualizerConfig `toml:"visualizer,omitempty"`
}

// SnakeAnimationConfig is the configuration for the snake animation.
type SnakeAnimationConfig struct {
	// Chunks is the list of chunks for the snake animation.
	Chunks []SnakeAnimationChunk `toml:"chunk"`
	// Speed is the speed of the snake animation.
	Speed TOMLDuration `toml:"speed"`
}

// SnakeAnimationChunk is a chunk for the snake animation.
type SnakeAnimationChunk struct {
	// Color is the color for the chunk.
	Color led.RGBColor `toml:"color"`
}

// VisualizerConfig is the configuration for the visualizer.
type VisualizerConfig struct {
	Kind    VisualizerKind `toml:"kind"`
	Flip    bool           `toml:"flip"`
	Bins    int            `toml:"bins"`
	Backend string         `toml:"backend"`
	Device  string         `toml:"device"`
	Smooth  float64        `toml:"smooth"`

	Gradients          []led.RGBColor `toml:"gradients"`
	GradientMode       GradientMode   `toml:"gradient_mode"`
	GradientPeakSwitch float64        `toml:"gradient_peak_switch"`
	GradientPeakBin    int            `toml:"gradient_peak_bin"`
	GradientDuration   TOMLDuration   `toml:"gradient_duration"`
}

// VisualizerKind is the kind of visualizer to use.
type VisualizerKind string

const (
	// GlowingVisualizer means glow each LED based on its frequency bin.
	GlowingVisualizer VisualizerKind = "glowing"
	// BlinkingVisualizer means glowing the entire strip based on the normalized
	// amplitude.
	BlinkingVisualizer VisualizerKind = "blinking"
	// MeterVisualizer shows a horizontal meter based on the normalized
	// amplitude.
	MeterVisualizer VisualizerKind = "meter"
)

// GradientMode is the mode for the gradient.
type GradientMode string

const (
	// PeakGradientMode means the gradient is changed based on the peak bin.
	PeakGradientMode GradientMode = "peak"
	// DurationGradientMode means the gradient is changed based on the duration.
	DurationGradientMode GradientMode = "duration"
	// StaticGradientMode means to only use the first color in the gradient.
	StaticGradientMode GradientMode = "static"
)

// TOMLDuration is a duration that can be parsed from TOML.
type TOMLDuration time.Duration

var (
	_ encoding.TextUnmarshaler = (*TOMLDuration)(nil)
	_ encoding.TextMarshaler   = (*TOMLDuration)(nil)
)

func (d *TOMLDuration) UnmarshalText(text []byte) error {
	duration, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}
	*d = TOMLDuration(duration)
	return nil
}

func (d TOMLDuration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

// ParseConfig parses a configuration from a reader.
func ParseConfig(r io.Reader) (*Config, error) {
	var config Config
	if err := toml.NewDecoder(r).Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}
