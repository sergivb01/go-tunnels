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

func PacketReader(reader BytesReader) (packetID uint64, packetReader BytesReader, err error) {
	length, err := readVarInt(reader)
	if err != nil {
		return
	}
	packet, err := ReadBytes(reader, length)
	if err != nil {
		return
	}
	packetReader = bytes.NewReader(packet)
	packetID, err = readVarInt(packetReader)
	return
}

func (stream *MCConn) ReadByte() (byte, error) {
	b, err := ReadBytes(stream, 1)
	return b[0], err
}

func ReadBytes(reader io.Reader, n uint64) (bytes []byte, err error) {
	bytes = make([]byte, n)
	_, err = reader.Read(bytes)
	return
}

func readVarInt(reader io.ByteReader) (num uint64, err error) {
	varint, err := binary.ReadUvarint(reader)
	if err != nil {
		return
	}
	return varint, err
}

func readShort(reader io.Reader) (num uint16, err error) {
	// 2 bytes required for short
	buffer, err := ReadBytes(reader, 2)
	if err != nil {
		return
	}
	return binary.BigEndian.Uint16(buffer), err
}

func readString(reader BytesReader) (result string, err error) {
	length, err := readVarInt(reader)
	if err != nil {
		return
	}
	buffer, err := ReadBytes(reader, length)
	if err != nil {
		return
	}
	return string(buffer), err
}
