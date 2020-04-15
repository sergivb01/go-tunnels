package packet

import (
	"io"
)

const (
	PACKET_HANDSHAKE = 0x00
)

type Packet interface {
	Encode(writer io.Writer) error
	Decode(reader io.Reader) error
	ID() int
}
