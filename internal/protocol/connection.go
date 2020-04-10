package protocol

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	"github.com/sergivb01/mctunnel/internal/protocol/packet"
	"github.com/sergivb01/mctunnel/internal/protocol/util"
)

var (
	ErrTooFewBytes = errors.New("too few bytes read")
)

// Connection represents a uni-directional connection from client to server.
type Connection struct {
	D  *Dealer
	rw io.ReadWriter
}

// NewConnection serves as the builder function for type Connection. It takes in
// a reader which, when read from, yeilds data sent by the "client".
func NewConnection(rw io.ReadWriter) *Connection {
	return &Connection{D: NewDealer(), rw: rw}
}

func (c *Connection) SetState(state State) {
	c.D.SetState(state)
}

func (c *Connection) Next() (*packet.Packet, error) {
	return c.packet()
}

func (c *Connection) Write(h packet.Holder) (int, error) {
	data, err := c.D.Encode(h)
	if err != nil {
		return -1, nil
	}

	return c.rw.Write(data)
}

// Next reads and decodes the next Packet on the stream. Packets are expected to
// be in the following format (as described on
// http://wiki.vg/Protocol#Without_compression:
//
// Without compression:
//   | Field Name | Field Type | Field Notes                        |
//   | ---------- | ---------- | ---------------------------------- |
//   | Length     | Uvarint    | Represents length of <id> + <data> |
//   | ID         | Uvarint    |                                    |
//   | Data       | []byte     |                                    |
//
// With compression:
// ...
//
// If an error is experienced in reading the packet from the io.Reader `r`, then
// a nil pointer will be returned and the error will be propogated up.
func (c *Connection) packet() (*packet.Packet, error) {
	r := bufio.NewReader(c.rw)

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
		return nil, ErrTooFewBytes
	}

	id, offset := binary.Uvarint(buffer)

	return &packet.Packet{
		ID:        int(id),
		Direction: packet.DirectionServerbound,
		Data:      buffer[offset:],
	}, nil
}

func (c *Connection) SendPacket(p *packet.Packet) {
	out := new(bytes.Buffer)
	out.Write(util.Uvarint(uint32(p.ID)))
	out.Write(util.Uvarint(uint32(len(p.Data))))
	out.Write(p.Data)
}

