package proto

import (
	"fmt"
	"io"
	"sync"

	"github.com/sergivb01/mctunnel/internal/proto/packet"
)

// PacketCodec manages the (de)code-ing of Minecraft packets
type PacketCodec struct {
	bpool *BufferPool
}

// NewPacketCodec returns a new PacketCodec
func NewPacketCodec() PacketCodec {
	return PacketCodec{
		bpool: &BufferPool{
			pool: &sync.Pool{
				New: NewBuffer,
			},
		},
	}
}

// EncodePacket encodes a Packet into the writer using the format (see comment from DecodePacket)
func (d *PacketCodec) EncodePacket(w io.Writer, p packet.Packet) error {
	buff := d.bpool.Get()
	defer d.bpool.Put(buff)

	if err := p.Encode(buff); err != nil {
		return fmt.Errorf("error encoding Packet to buffer: %w", err)
	}

	if err := packet.WriteVarInt(w, buff.Len()); err != nil {
		return fmt.Errorf("error encoding buffer length: %w", err)
	}

	if err := packet.WriteVarInt(w, p.ID()); err != nil {
		return fmt.Errorf("error encoding PacketID: %w", err)
	}

	if _, err := w.Write(buff.Bytes()); err != nil {
		return fmt.Errorf("error writing buffer to Writer: %w", err)
	}

	return nil
}

// DecodePacket reads and decodes the next Packet size and ID on the stream. Packets are expected to
// be in the following format, as described on
// http://wiki.vg/Protocol#Without_compression:
//
// Without compression:
//   | Field Name | Field Type | Field Notes                        |
//   | ---------- | ---------- | ---------------------------------- |
//   | Length     | Uvarint    | Represents length of <id> + <data> |
//   | ID         | Uvarint    |                                    |
//   | Data       | []byte     |                                    |
//
// If an error is experienced in reading the packet from the io.Reader `r`, then
// a nil pointer will be returned and the error will be propagated up.
// TODO: if size is too small we could try and not read the packet as it will probably be invalid?
// TODO: also decode packet and return it, then use pk.(*PacketType) to handle it where necessary
func (d *PacketCodec) DecodePacket(r io.Reader) (int, error) {
	// packet length, we don't care about the size but we need to read those bytes
	if _, err := packet.ReadVarInt(r); err != nil {
		return 0, err
	}
	return packet.ReadVarInt(r) // packetID
}
