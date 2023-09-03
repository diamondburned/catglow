# catglow

ESP32-based music visualizer for WS2812 LED strips.

## Working Mechanism

This project uses [noriah/catnip](https://github.com/noriah/catnip) to process
audio on the system into frequency bins, and then uses those frequency bins to
render a visualization to be sent to the ESP32, which acts as a controller for
a WS2812 LED strip.

### Diagram

```
┌───────────────────────────────────────────────────────────────────┐
│                                                                   │
│   Linux Machine                                                   │
│                      ┌──────────────────────┐                     │
│                      │                      │                     │   ┌─────────────┐
│                      │  ┌────────────────┐  │  ┌──────────────┐   │   │             │
│                      │  │                │  │  │              │   │   │             │
│                      │  │  LED Renderer  ├──┼──► /dev/ttyUSB0 ├───┼───►    ESP32    │
│                      │  │                │  │  │              │   │   │             │
│                      │  └───────▲────────┘  │  └──────────────┘   │   │ <ledserial> │
│   ┌───────────────┐  │          │           │                     │   │             │
│   │ Audio Players │  │  ┌───────┴────────┐  │                     │   │             │
│   │               │  │  │                │  │                     │   └──────┬──────┘
│   │ Spotify       ├──┼──►     catnip     │  │                     │          │
│   │ Firefox       │  │  │                │  │                     │          │
│   │ Mixxx         │  │  └────────────────┘  │                     │   ┌──────▼───────
│   │ ...           │  │                      │                     │   │
│   └───────────────┘  └──────────────────────┘                     │   │    LEDs...
│                                                                   │   │
└───────────────────────────────────────────────────────────────────┘   └──────────────
```

## Build

```sh
go build
```

## Usage

```sh
./catglow -c catglow.toml # run with a config file
```

## Configuration

```toml
[[led]]
  range = [40, 192]

  visualizer = "glowing" # see #Visualizers
  backend = "pipewire"
  device = "spotify"
  smooth = 0.5
  flip = true # flip the LED strip

  [led.visualizer]
   bins = -1 # use as many bins as there are LEDs, otherwise bins are sectioned

   gradients = [
     [255, 0, 0],
     [0, 255, 0],
     [0, 0, 255],
   ]
   gradient_mode = "peak"      # "peak" or "duration" or "static"
   gradient_peak_switch = 0.85 # switch to the next gradient when the peak is above 85%
   gradient_peak_bin = 0       # use the first frequency bin for the peak
   gradient_duration = "1s"    # used if gradient_mode is "duration"

[[led]]
  range = [0, 40]
  color = [255, 255, 255] # static color
```

## Visualizers

- `glowing`: glow each LED based on the frequency bin.
- `blinking`: blink the entire LED strip based on the normalized amplitude.
- `meter`: show a horizontal meter based on the normalized amplitude.
