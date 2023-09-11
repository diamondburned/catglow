package ledvis

import "github.com/noriah/catnip/processor"

// Blinking is a visualization that blinks the LEDs based on the normalized
// amplitude of the audio.
type Blinking struct {
	cfg VisualizerConfig
}

func NewBlinking(cfg VisualizerConfig) (*Blinking, error) {

}

type blinkingOutput struct {
	*Blinking
}

var _ processor.Output = (*blinkingOutput)(nil)

func (o blinkingOutput) Bins() int {
	return o.cfg.Bins
}

func (o blinkingOutput) Write(bins [][]float64, nchannels int) error {
	return nil
}
