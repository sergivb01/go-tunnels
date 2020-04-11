package proto

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
)

const (
	maxVarIntLength = 5
	maxShortLength  = 2
)

type MCConn struct {
	*net.TCPConn
}

func packetWriter(packetID uint64) (writer *bytes.Buffer, err error) {
	writer = new(bytes.Buffer)
	err = writeVarInt(writer, packetID)
	return
}

func encodeVarInt(num uint64) []byte {
	buffer := make([]byte, maxVarIntLength)
	return buffer[:binary.PutUvarint(buffer, num)]
}

func writeVarInt(writer io.Writer, num uint64) (err error) {
	_, err = writer.Write(encodeVarInt(num))
	return
}

func writeShort(writer io.Writer, num uint16) error {
	_, err := writer.Write(encodeShort(num))
	return err
}

func encodeShort(num uint16) []byte {
	buffer := make([]byte, maxShortLength)
	binary.BigEndian.PutUint16(buffer, num)
	return buffer
}

func writeString(writer io.Writer, s string) (err error) {
	err = writeVarInt(writer, uint64(len(s)))
	if err != nil {
		return
	}
	_, err = io.WriteString(writer, s)
	return
}
