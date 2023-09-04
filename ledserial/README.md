Package ledserial implements a program that reads LED colors from the UART
serial input and displays them on the LED strip.

# Protocol

The protocol defines a simple byte protocol for sending various packet types
over the wire.

## Scalar Types

The following scalar types are defined:

    - uint8:   1 byte unsigned integer
    - uint16:  2 byte unsigned integer, little endian
    - uint32:  4 byte unsigned integer, little endian
    - int8:    1 byte signed integer
    - int16:   2 byte signed integer, little endian
    - int32:   4 byte signed integer, little endian
    - float32: 4 byte IEEE 754 floating point number, little endian
    - string:  uint16 length followed by a UTF-8 encoded string of the given
               length
    - bytes:   uint16 length followed by the given number of bytes

# Incoming Packet

Each packet starts with a single byte that defines the packet type. The
following packet types are defined:

	0x00: Initialize packet. This must be sent before any other packet.
	0x01: Clear all LEDs.
	0x02: Set all LEDs to the given colors.

All packets must be suffixed with a CRC32 checksum with the IEEE polynomial.
The checksum is calculated over the entire packet, including the packet type.
The checksum is sent as a uint32.

The packet structure can be understood as follows:

	0x00: Packet type (uint8)
	0x01: Packet data of known length depending on the packet type
    0xNN: CRC32 checksum (uint32)

The packet data depends on the packet type.

## Initialize Packet

The initialize packet is sent as a single byte with value 0x00. It resets the
LED strip to its initial state. The packet requires the following data:

	0x00: 0x01 value (uint8)
	0x01: Number of LEDs (uint16)

## Clear Packet

The clear packet is sent as a single byte with value 0x01. It clears all LEDs
to black. The packet requires no additional data.

## Set Packet

The set packet is sent as a single byte with value 0x02. It sets all LEDs to
the given colors. The packet requires the following data:

	0x00: 0x02 value (uint8)
	0x01: Red value of the first LED (uint8)
	0x02: Green value of the first LED (uint8)
	0x03: Blue value of the first LED (uint8)
	0x04: Red value of the second LED (uint8)
	0x05: Green value of the second LED (uint8)
	0x06: Blue value of the second LED (uint8)
	...

The total number of LEDs must be equal to the number of LEDs specified in the
initialize packet. The total length of the packet would be `3*numLEDs + 1`.

To better visualize the packet structure, here is a diagram of the packet
structure:

	+--------+--------+--------+--------+--------+--------+--------+--------+
	+ Type   | Data                                                         |
	+--------+--------+--------+--------+--------+--------+--------+--------+
	| 0x02   | LED 0                    | LED 1                    | ...	|
	+        +                          +                          +        +
	|        | Red    | Green  | Blue   | Red    | Green  | Blue   | ...	|
	+--------+--------+--------+--------+--------+--------+--------+--------+

# Outgoing Packet

Each packet starts with a single byte that defines the packet type. The
following packet types are defined:

	0x00: Error packet. This is sent when an error occurs.
	0x01: Panic packet. This is sent when the program cannot recover.
    0x03: Log packet. This is sent when the program wants to log a message.
    0x04: Acknowledgement packet. This is sent when a packet is received.

The packet structure can be similarly understood as the incoming packet.

## Error Packet

The error packet is sent as a single byte with value 0x00. It indicates that
an error occurred. The packet requires the following data:

	0x00: 0x00 value (uint8)
	0x01: Error message (string)

## Panic Packet

The panic packet is sent as a single byte with value 0x01. It indicates that
the program cannot recover. The packet requires the following data:

	0x00: 0x01 value (uint8)

## Log Packet

The log packet is sent as a single byte with value 0x03. It indicates that
the program wants to log a message. The packet requires the following data:

    0x00: 0x03 value (uint8)
    0x01: Log message (string)
