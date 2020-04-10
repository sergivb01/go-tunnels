package mcproto

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

const maxVarintLength = 5

var ByteOrder = binary.BigEndian

func (h *Handshake) EncodePacket(addr string) ([]byte, error) {
	out := new(bytes.Buffer)
	out.Write(Uvarint(PacketIdHandshake))

	if _, err := WriteUVarInt(out, uint32(h.ProtocolVersion)); err != nil {
		return nil, fmt.Errorf("error writing protover: %w", err)
	}

	if err := WriteString(out, addr); err != nil {
		return nil, fmt.Errorf("error writing srv address: %w", err)
	}

	if _, err := WriteUShort(out, h.ServerPort); err != nil {
		return nil, fmt.Errorf("error writing server port: %w", err)
	}

	if _, err := WriteUVarInt(out, uint32(h.NextState)); err != nil {
		return nil, fmt.Errorf("error writing next state: %w", err)
	}

	return append(
		Uvarint(uint32(out.Len())),
		out.Bytes()...,
	), nil
}

// TODO: proper encode packet (this is really bad lol)
func (h *LegacyServerListPing) BadEncoding(addr string) ([]byte, error) {
	out := new(bytes.Buffer)
	out.Write(Uvarint(PacketIdLegacyServerListPing))

	if _, err := WriteUVarInt(out, uint32(h.ProtocolVersion)); err != nil {
		return nil, fmt.Errorf("error writing protover: %w", err)
	}

	if err := WriteString(out, addr); err != nil {
		return nil, fmt.Errorf("error writing srv address: %w", err)
	}

	if _, err := WriteUShort(out, h.ServerPort); err != nil {
		return nil, fmt.Errorf("error writing server port: %w", err)
	}

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
