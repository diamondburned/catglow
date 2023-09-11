package ledvis

import (
	"sync"

	"libdb.so/catglow/internal/led"
)

type baseOutput struct {
	mu   sync.Mutex
	leds led.LEDs
}

func (o *baseOutput) AcquireFrame(f func(led.LEDs)) {
	o.mu.Lock()
	f(o.leds)
	o.mu.Unlock()
}
