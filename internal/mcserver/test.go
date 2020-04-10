package mcserver

import (
	"bufio"
	"encoding/binary"
	"io"

	"github.com/sergivb01/mctunnel/internal/protocol"
	"github.com/sergivb01/mctunnel/internal/protocol/packet"
)

func testDecode(reader io.ReadWriter) (*packet.Packet, error) {
	r := bufio.NewReader(reader)

	size, err := binary.ReadUvarint(r)
	if err != nil {
		return nil, err
	}

	// TODO(ttaylorr): extract this to a package `util`
	buffer := make([]byte, size)
	read, err := io.ReadAtLeast(r, buffer, int(size))
	if err != nil {
		return nil, err
	} else if read < int(size) {
		return nil, protocol.ErrTooFewBytes
	}

	id, offset := binary.Uvarint(buffer)

	return &packet.Packet{
		ID:        int(id),
		Direction: packet.DirectionServerbound,
		Data:      buffer[offset:],
	}, nil
}
