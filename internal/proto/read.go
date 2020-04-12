package proto

import (
	"bytes"
	"encoding/binary"
	"io"
)

type BytesReader interface {
	io.Reader
	io.ByteReader
}

func PacketReader(reader BytesReader) (uint64, BytesReader, error) {
	length, err := readVarInt(reader)
	if err != nil {
		return 0, nil, err
	}
	packet, err := ReadBytes(reader, length)
	if err != nil {
		return 0, nil, err
	}
	packetReader := bytes.NewReader(packet)
	packetID, err := readVarInt(packetReader)
	return packetID, packetReader, err
}

func (stream *MCConn) ReadByte() (byte, error) {
	b, err := ReadBytes(stream, 1)
	return b[0], err
}

func ReadBytes(reader io.Reader, n uint64) ([]byte, error) {
	b := make([]byte, n)
	_, err := reader.Read(b)
	return b, err
}

func readVarInt(reader io.ByteReader) (uint64, error) {
	return binary.ReadUvarint(reader)
}

func readShort(reader io.Reader) (uint16, error) {
	// 2 bytes required for short
	buffer, err := ReadBytes(reader, 2)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(buffer), err
}

func readString(reader BytesReader) (string, error) {
	length, err := readVarInt(reader)
	if err != nil {
		return "", err
	}
	buffer, err := ReadBytes(reader, length)
	if err != nil {
		return "", err
	}
	return string(buffer), err
}
