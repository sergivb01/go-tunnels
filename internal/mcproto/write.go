package mcproto

import (
	"bytes"
	"encoding/binary"
	"io"
)

const maxVarintLength = 5

var ByteOrder = binary.BigEndian

func (h *Handshake) EncodePacket(addr string) ([]byte, error) {
	out := new(bytes.Buffer)
	out.Write(Uvarint(0))

	// 	ProtocolVersion types.UVarint
	//	ServerAddress   types.String
	//	ServerPort      types.UShort
	//	NextState       types.UVarint
	WriteUVarInt(out, uint32(h.ProtocolVersion))
	WriteString(out, addr)
	WriteUShort(out, 25565)
	WriteUVarInt(out, 2)

	return append(
		Uvarint(uint32(out.Len())),
		out.Bytes()...,
	), nil
}


func (h *LegacyServerListPing) EncodePacket(addr string) ([]byte, error) {
	out := new(bytes.Buffer)
	out.Write(Uvarint(0))

	// 	ProtocolVersion types.UVarint
	//	ServerAddress   types.String
	//	ServerPort      types.UShort
	//	NextState       types.UVarint
	WriteUVarInt(out, uint32(h.ProtocolVersion))
	WriteString(out, addr)
	WriteUShort(out, 25565)

	return append(
		Uvarint(uint32(out.Len())),
		out.Bytes()...,
	), nil
}

func WriteLong(w io.Writer, i int64) error {
	return binary.Write(w, ByteOrder, i)
}

func WriteUShort(w io.Writer, i uint16) (int, error) {
	buf := make([]byte, 2)
	ByteOrder.PutUint16(buf, i)

	return w.Write(buf)
}

func WriteUVarInt(w io.Writer, i uint32) (int, error) {
	return w.Write(Uvarint(i))
}

func WriteVarInt(w io.Writer, i int32) (int, error) {
	return w.Write(Varint(i))
}

func WriteString(w io.Writer, s string) error {
	var n int
	var err error

	length := Uvarint(uint32(len(s)))

	written, err := w.Write(length)
	n += written
	if err != nil {
		return err
	}

	written, err = w.Write([]byte(s))
	n += written
	if err != nil {
		return err
	}

	return nil
}
