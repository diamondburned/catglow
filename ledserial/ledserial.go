// Package ledserial implements the LED serial protocol.
package ledserial

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
)

// Endianness defines the endianness of the protocol.
var Endianness = binary.LittleEndian

// IncomingPacketType is a type of packet.
type IncomingPacketType uint8

const (
	TypeInitializePacket IncomingPacketType = iota
	TypeClearPacket
	TypeSetPacket
)

// String returns a string representation of the packet type.
func (t IncomingPacketType) String() string {
	switch t {
	case TypeInitializePacket:
		return "initialize"
	case TypeClearPacket:
		return "clear"
	case TypeSetPacket:
		return "set"
	default:
		return fmt.Sprintf("IncomingPacketType(%d)", t)
	}
}

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
	TypeLogPacket
)

// String returns a string representation of the packet type.
func (t OutgoingPacketType) String() string {
	switch t {
	case TypeErrorPacket:
		return "error"
	case TypePanicPacket:
		return "panic"
	case TypeLogPacket:
		return "log"
	default:
		return fmt.Sprintf("OutgoingPacketType(%d)", t)
	}
}

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

// LogPacket is a packet that contains a log message.
type LogPacket struct {
	Message string
}

func (p ErrorPacket) Type() OutgoingPacketType { return TypeErrorPacket }
func (p PanicPacket) Type() OutgoingPacketType { return TypePanicPacket }
func (p LogPacket) Type() OutgoingPacketType   { return TypeLogPacket }

// Reader is a reader that reads packets.
type Reader interface {
	io.ByteReader
	io.Reader
}

// ReadContext is the state of the LED strip. Data in this structure are
// required for the device to read incoming packets.
type ReadContext struct {
	// NumLEDs is the number of LEDs in the strip.
	NumLEDs uint16
}

// ReadIncomingPacket reads an incoming packet from the given reader.
func ReadIncomingPacket(r io.Reader, context ReadContext) (IncomingPacket, error) {
	hash := crc32.NewIEEE()
	r = io.TeeReader(r, hash)

	var packet IncomingPacket
	var ptypeBuf [1]byte
	if _, err := io.ReadFull(r, ptypeBuf[:]); err != nil {
		return nil, fmt.Errorf("failed to read incoming packet type: %w", err)
	}

	switch ptype := IncomingPacketType(ptypeBuf[0]); ptype {
	case TypeInitializePacket:
		var p InitializePacket
		if err := binary.Read(r, Endianness, &p); err != nil {
			return nil, fmt.Errorf("failed to read number of LEDs: %w", err)
		}
		packet = p

	case TypeClearPacket:
		var p ClearPacket
		packet = p

	case TypeSetPacket:
		var p SetPacket
		p.Pix = make([]uint8, 3*context.NumLEDs)
		if _, err := io.ReadFull(r, p.Pix); err != nil {
			return nil, fmt.Errorf("failed to read pixel data: %w", err)
		}
		packet = p

	default:
		return nil, fmt.Errorf("unknown packet type: %s", ptype)
	}

	var checksum uint8
	if err := binary.Read(r, Endianness, &checksum); err != nil {
		return nil, fmt.Errorf("failed to read checksum: %w", err)
	}

	if checksum != uint8(hash.Sum32()) {
		return nil, fmt.Errorf("checksum mismatch")
	}

	return packet, nil
}

// WriteIncomingPacket writes an incoming packet to the given writer.
func WriteIncomingPacket(w io.Writer, p IncomingPacket) error {
	hash := crc32.NewIEEE()
	w = io.MultiWriter(w, hash)

	switch p := p.(type) {
	case InitializePacket:
		if err := binary.Write(w, Endianness, TypeInitializePacket); err != nil {
			return fmt.Errorf("failed to write packet type: %w", err)
		}
		if err := binary.Write(w, Endianness, p); err != nil {
			return fmt.Errorf("failed to write packet: %w", err)
		}
	case ClearPacket:
		if err := binary.Write(w, Endianness, TypeClearPacket); err != nil {
			return fmt.Errorf("failed to write packet type: %w", err)
		}
	case SetPacket:
		if err := binary.Write(w, Endianness, TypeSetPacket); err != nil {
			return fmt.Errorf("failed to write packet type: %w", err)
		}
		if _, err := w.Write(p.Pix); err != nil {
			return fmt.Errorf("failed to write packet: %w", err)
		}
	default:
		return fmt.Errorf("unknown packet type: %T", p)
	}

	if err := binary.Write(w, Endianness, hash.Sum32()); err != nil {
		return fmt.Errorf("failed to write packet checksum: %w", err)
	}

	return nil
}

// ReadOutgoingPacket reads an outgoing packet from the given reader.
func ReadOutgoingPacket(r io.Reader, context ReadContext) (OutgoingPacket, error) {
	hash := crc32.NewIEEE()
	r = io.TeeReader(r, hash)

	var packet OutgoingPacket
	var ptypeBuf [1]byte
	if _, err := io.ReadFull(r, ptypeBuf[:]); err != nil {
		return nil, fmt.Errorf("failed to read outgoing packet type: %w", err)
	}

	switch ptype := OutgoingPacketType(ptypeBuf[0]); ptype {
	case TypeErrorPacket:
		var p ErrorPacket
		var length uint16
		if err := binary.Read(r, Endianness, &length); err != nil {
			return nil, fmt.Errorf("failed to read error message length: %w", err)
		}
		buf := make([]byte, length)
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, fmt.Errorf("failed to read error message: %w", err)
		}
		p.Message = string(buf)
		packet = p

	case TypePanicPacket:
		var p PanicPacket
		packet = p

	case TypeLogPacket:
		var p LogPacket
		var length uint16
		if err := binary.Read(r, Endianness, &length); err != nil {
			return nil, fmt.Errorf("failed to read log message length: %w", err)
		}
		buf := make([]byte, length)
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, fmt.Errorf("failed to read log message: %w", err)
		}
		p.Message = string(buf)
		packet = p

	default:
		return nil, fmt.Errorf("unknown packet type: %s", ptype)
	}

	var checksum uint32
	if err := binary.Read(r, Endianness, &checksum); err != nil {
		return nil, fmt.Errorf("failed to read packet checksum: %w", err)
	}

	if checksum != hash.Sum32() {
		return nil, fmt.Errorf("packet checksum mismatch")
	}

	return packet, nil
}

// WriteOutgoingPacket writes an outgoing packet to the given writer.
func WriteOutgoingPacket(w io.Writer, p OutgoingPacket) error {
	hash := crc32.NewIEEE()
	w = io.MultiWriter(w, hash)

	switch p := p.(type) {
	case ErrorPacket:
		if err := binary.Write(w, Endianness, TypeErrorPacket); err != nil {
			return fmt.Errorf("failed to write packet type: %w", err)
		}
		if err := binary.Write(w, Endianness, uint16(len(p.Message))); err != nil {
			return fmt.Errorf("failed to write error message length: %w", err)
		}
		if _, err := w.Write([]byte(p.Message)); err != nil {
			return fmt.Errorf("failed to write error message: %w", err)
		}
	case PanicPacket:
		if err := binary.Write(w, Endianness, TypePanicPacket); err != nil {
			return fmt.Errorf("failed to write packet type: %w", err)
		}
	case LogPacket:
		if err := binary.Write(w, Endianness, TypeLogPacket); err != nil {
			return fmt.Errorf("failed to write packet type: %w", err)
		}
		if err := binary.Write(w, Endianness, uint16(len(p.Message))); err != nil {
			return fmt.Errorf("failed to write log message length: %w", err)
		}
		if _, err := w.Write([]byte(p.Message)); err != nil {
			return fmt.Errorf("failed to write log message: %w", err)
		}
	default:
		return fmt.Errorf("unknown packet type: %T", p)
	}

	if err := binary.Write(w, Endianness, hash.Sum32()); err != nil {
		return fmt.Errorf("failed to write packet checksum: %w", err)
	}

	return nil
}
