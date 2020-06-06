package proto

import (
	"fmt"
	"io"
	"sync"

	"github.com/minebreach/go-tunnels/internal/proto/packet"
)

// PacketCodec manages the (de)code-ing of Minecraft packets
type PacketCodec struct {
	bpool *BufferPool
}

// NewPacketCodec returns a new PacketCodec
func NewPacketCodec() *PacketCodec {
	return &PacketCodec{
		bpool: &BufferPool{
			pool: &sync.Pool{
				New: NewBuffer,
			},
		},
	}
}

// WritePacket encodes a Packet into the writer using the format (see comment from ReadPacket)
func (d *PacketCodec) WritePacket(w io.Writer, p packet.Packet) error {
	buff := d.bpool.Get()
	defer d.bpool.Put(buff)

	// write the PacketID to the buffer of data
	if err := packet.WriteVarInt(buff, p.ID()); err != nil {
		return fmt.Errorf("encoding PacketID: %w", err)
	}

	// write the packet data
	if err := p.Encode(buff); err != nil {
		return fmt.Errorf("encoding Packet to buffer: %w", err)
	}

	// write the buffer length
	if err := packet.WriteVarInt(w, buff.Len()); err != nil {
		return fmt.Errorf("encoding buffer length: %w", err)
	}

	// write the buffer -> len (ID+data), ID, data
	if _, err := w.Write(buff.Bytes()); err != nil {
		return fmt.Errorf("writing buffer to Writer: %w", err)
	}

	return nil
}

// ReadPacket reads a Packet following the specified format in and returns the PacketID and error
func (d *PacketCodec) ReadPacket(r io.Reader) (int, error) {
	if _, err := packet.ReadVarInt(r); err != nil {
		return 0, fmt.Errorf("reading length: %w", err)
	}

	packetID, err := packet.ReadVarInt(r)
	if err != nil {
		return 0, fmt.Errorf("reading packetid: %w", err)
	}

	return packetID, nil
}
