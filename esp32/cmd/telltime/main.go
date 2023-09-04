package main

import (
	"fmt"
	"io"
	"machine"
	"time"
)

func main() {
	for t := range time.Tick(time.Second) {
		io.WriteString(machine.Serial, fmt.Sprintf("the time is now %s\n", t))
	}
}
