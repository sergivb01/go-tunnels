package packet

import (
	"io"
)

const (
	HandshakeId = 0x00
)

type Packet interface {
	// Encode encodes a Packet into a io.Writer
	Encode(writer io.Writer) error

	// Decode decodes a Packet from an io.Reader
	Decode(reader io.Reader) error

	// ID returns the ID of the Packet
	ID() int
}
