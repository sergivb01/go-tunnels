package packet

import (
	"io"
)

const (
	// HandshakeID is the ID from the Handshaking state
	HandshakeID = 0x00
)

// Packet specifies the interface for a Minecraft Packet
type Packet interface {
	// Encode encodes a Packet into a io.Writer
	Encode(writer io.Writer) error

	// Decode decodes a Packet from an io.Reader
	Decode(reader io.Reader) error

	// ID returns the ID of the Packet
	ID() int
}
