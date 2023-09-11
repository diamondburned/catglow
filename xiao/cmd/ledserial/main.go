package main

import "machine"

func main() {
	device := NewDevice(machine.Serial, machine.D0)
	device.Run()
}
