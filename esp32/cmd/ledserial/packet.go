package main

// IncomingPacketType is a type of packet.
type IncomingPacketType uint8

const (
	TypeInitializePacket IncomingPacketType = iota
	TypeClearPacket
	TypeSetPacket
)

// IncomingPacket is a packet sent over the wire.
type IncomingPacket interface {
	// Type returns the type of packet.
	Type() IncomingPacketType
}

// InitializePacket is a packet that initializes the LED strip.
type InitializePacket struct {
	NumLEDs uint16
}

// ClearPacket is a packet that clears the LED strip.
type ClearPacket struct{}

// SetPacket is a packet that sets the LED strip to the given colors.
type SetPacket struct {
	Pix []uint8
}

func (p InitializePacket) Type() IncomingPacketType { return TypeInitializePacket }
func (p ClearPacket) Type() IncomingPacketType      { return TypeClearPacket }
func (p SetPacket) Type() IncomingPacketType        { return TypeSetPacket }

// OutgoingPacketType is a type of packet.
type OutgoingPacketType uint8

const (
	TypeErrorPacket OutgoingPacketType = iota
	TypePanicPacket
)

// OutgoingPacket is a packet sent over the wire.
type OutgoingPacket interface {
	// Type returns the type of packet.
	Type() OutgoingPacketType
}

// ErrorPacket is a packet that indicates an error occurred.
type ErrorPacket struct {
	Message string
}

// PanicPacket is a packet that indicates the program cannot recover.
type PanicPacket struct{}

func (p ErrorPacket) Type() OutgoingPacketType { return TypeErrorPacket }
func (p PanicPacket) Type() OutgoingPacketType { return TypePanicPacket }
