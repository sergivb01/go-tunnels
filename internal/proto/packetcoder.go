package proto

import (
	"io"

	"github.com/sergivb01/mctunnel/internal/proto/packet"
)

type PacketCoder struct {
	bpool *BufferPool
}

func NewPacketDecoder() PacketCoder {
	return PacketCoder{bpool: NewBufferPool()}
}

func (d *PacketCoder) EncodePacket(w io.Writer, p packet.Packet) error {
	buff := d.bpool.Get()
	defer d.bpool.Put(buff)

	if err := p.Encode(buff); err != nil {
		return err
	}

	if err := packet.WriteVarInt(w, buff.Len()); err != nil {
		return err
	}

	if err := packet.WriteVarInt(w, p.ID()); err != nil {
		return err
	}

	_, err := w.Write(buff.Bytes())
	return err
}

func (d *PacketCoder) DecodePacket(r io.Reader) (int, error) {
	// TODO: if size is too small we could try and not read the packet as it will probably be invalid?
	if _, err := packet.ReadVarInt(r); err != nil {
		return 0, err
	}
	return packet.ReadVarInt(r)
}
