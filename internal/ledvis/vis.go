package ledvis

// ChannelStyle is the style to draw the channels in.
type ChannelStyle uint8

const (
	// MonoLeft means that a single mono channel is drawn from the left.
	MonoLeft ChannelStyle = iota
	// MonoRight means that a single mono channel is drawn from the right.
	MonoRight
	// StereoTypeSymmetricMiddle means that the left and right channels are
	// drawn symmetrically from the middle and outwards.
	StereoTypeSymmetricMiddle
)

// NumChannels returns the number of channels for the given channel style.
func (s ChannelStyle) NumChannels() int {
	switch s {
	case MonoLeft, MonoRight:
		return 1
	case StereoTypeSymmetricMiddle:
		return 2
	default:
		panic("invalid channel style")
	}
}

func (s ChannelStyle) String() string {
	switch s {
	case MonoLeft:
		return "mono-left"
	case MonoRight:
		return "mono-right"
	case StereoTypeSymmetricMiddle:
		return "stereo-symmetric-middle"
	default:
		panic("invalid channel style")
	}
}

// VisualizerConfig is the configuration for the visualizer.
type VisualizerConfig struct {
	Backend string
	Device  string
	// Bins is the number of bins to use for the visualizer.
	// If not set, then the number of LEDs is used.
	Bins int
	// SmoothFactor is the smooth factor to use for the visualizer.
	SmoothFactor float64
	// ChannelStyle is the channel style to use for the visualizer.
	ChannelStyle ChannelStyle
}
