package catglow

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/pkg/errors"
	"go.bug.st/serial"
	"golang.org/x/sync/errgroup"
	"libdb.so/catglow/ledserial"
)

// RefreshQueuer is the interface for types that can queue a refresh of the
// LEDs. LED animations use this interface to queue a refresh when they are
// done.
type RefreshQueuer interface {
	// QueueRefresh queues a refresh of the LEDs.
	// The daemon may choose to ignore this request if it is already refreshing
	// the LEDs.
	QueueRefresh()
}

// Animator is the interface for types that can animate the LEDs.
// It is kept to a minimum.
type Animator interface {
	// AcquireFrame acquires a frame from the animator. The frame is passed to
	// the callback function. The callback function must not be called after
	// AcquireFrame returns.
	// Usually, the daemon will call this method when QueueRefresh is called.
	AcquireFrame(f func(LEDs))
}

// Daemon is the main catglow daemon.
type Daemon struct {
	cfg     *Config
	logger  *slog.Logger
	refresh chan struct{}
}

var _ RefreshQueuer = (*Daemon)(nil)

// NewDaemon creates a new catglow daemon.
func NewDaemon(cfg *Config, logger *slog.Logger) (*Daemon, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid configuration")
	}

	return &Daemon{
		cfg:     cfg,
		logger:  logger,
		refresh: make(chan struct{}, 1),
	}, nil
}

// QueueRefresh queues a refresh of the LEDs.
// This method is mainly used internally.
func (d *Daemon) QueueRefresh() {
	select {
	case d.refresh <- struct{}{}:
	default:
	}
}

// Run starts the daemon. It blocks until the given context is canceled.
func (d *Daemon) Run(ctx context.Context) error {
	return (&internalDaemon{Daemon: d}).Run(ctx)
}

type trackedAnimator struct {
	Animator
	cfg LEDConfig
}

type internalDaemon struct {
	*Daemon
	port serial.Port
}

func (d *internalDaemon) Run(ctx context.Context) error {
	port, err := serial.Open(d.cfg.Device, &serial.Mode{
		BaudRate: d.cfg.Baud,
	})
	if err != nil {
		return errors.Wrap(err, "failed to open serial port")
	}
	defer port.Close()

	d.port = port
	if !d.writePacket(ctx, ledserial.InitializePacket{
		NumLEDs: uint16(d.cfg.NumLEDs()),
	}) {
		return errors.New("failed to initialize LEDs")
	}

	outPackets := make(chan ledserial.OutgoingPacket)

	errg, ctx := errgroup.WithContext(ctx)
	errg.Go(func() error {
		<-ctx.Done()
		d.logger.Debug("closing serial port")
		if err := port.Close(); err != nil {
			return errors.Wrap(err, "failed to close serial port")
		}
		return ctx.Err()
	})
	errg.Go(func() error {
		return d.mainLoop(ctx)
	})
	errg.Go(func() error {
		return d.readPackets(ctx, outPackets)
	})
	errg.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case p := <-outPackets:
				if err := d.handlePacket(p); err != nil {
					return err
				}
			}
		}
	})
	return errg.Wait()
}

func (d *internalDaemon) mainLoop(ctx context.Context) error {
	leds := NewLEDs(d.cfg.NumLEDs())
	var animators []trackedAnimator

	for _, led := range d.cfg.LEDs {
		switch {
		case led.Color != nil:
			// Pre-initialize with static colors and skip the animator.
			if led.Color != nil {
				leds.SetRange(led.Range[0], led.Range[1], *led.Color)
			}
		case led.Snake != nil:
		case led.Visualizer != nil:
		}
	}

	frameTicker := time.NewTicker(time.Second / time.Duration(d.cfg.Rate))
	defer frameTicker.Stop()

	// var nextFrame <-chan time.Time // nil unless invalidated
	// refresh := d.refresh           // nil when refresh is not done
	// d.QueueRefresh()

eventLoop:
	for {
		select {
		case <-ctx.Done():
			break eventLoop

		// case <-refresh:
		// 	nextFrame = frameTicker.C
		// 	refresh = nil

		case <-frameTicker.C:
			// nextFrame = nil
			// refresh = d.refresh

			for _, animator := range animators {
				animator.AcquireFrame(func(f LEDs) {
					f.Draw(animator.cfg.Range[0], leds)
				})
			}

			d.writePacket(ctx, ledserial.SetPacket{
				Pix: leds.AsPixels(),
			})
		}
	}

	return nil
}

func (d *internalDaemon) readPackets(ctx context.Context, dst chan<- ledserial.OutgoingPacket) error {
	if err := d.port.SetReadTimeout(serial.NoTimeout); err != nil {
		return errors.Wrap(err, "failed to reset read timeout")
	}

	for ctx.Err() == nil {
		p, err := ledserial.ReadOutgoingPacket(d.port, ledserial.ReadContext{
			NumLEDs: uint16(d.cfg.NumLEDs()),
		})
		if err != nil {
			// A short read indicates a timeout. This is expected.
			// Ignore the error and try again.
			if errors.Is(err, io.EOF) {
				continue
			}
			return errors.Wrap(err, "failed to read packet")
		}

		d.logger.Debug(
			"received packet from controller",
			"type", p.Type())

		select {
		case <-ctx.Done():
			return ctx.Err()
		case dst <- p:
			// ok
		}
	}

	return ctx.Err()
}

func (d *internalDaemon) handlePacket(p ledserial.OutgoingPacket) error {
	switch p := p.(type) {
	case ledserial.ErrorPacket:
		d.logger.Warn(
			"received error packet from controller",
			"message", p.Message)
		return nil

	case ledserial.PanicPacket:
		d.logger.Error("controller unrecoverably panicked")
		return errors.New("controller panicked")

	default:
		d.logger.Warn(
			"received unknown packet from controller",
			"packet", p.Type())
		return nil
	}
}

func (d *internalDaemon) writePacket(ctx context.Context, p ledserial.IncomingPacket) bool {
	d.logger.Debug(
		"writing packet",
		"type", p.Type())

	if err := ledserial.WriteIncomingPacket(d.port, p); err != nil {
		d.logger.Warn(
			"failed to write packet",
			"packet", p.Type(),
			"error", err)
		return false
	}

	if err := d.port.Drain(); err != nil {
		d.logger.Warn(
			"failed to drain serial port",
			"error", err)
		return false
	}

	return true
}