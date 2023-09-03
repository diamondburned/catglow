package main

import "machine"

func main() {
	device := NewDevice(machine.Serial, machine.GPIO27)
	device.Run()
}
