package packet

import (
	"io"
)

type Handshake struct {
	ProtocolVersion int
	ServerAddress   string
	ServerPort      uint16
	State           int
}

func (h *Handshake) Encode(writer io.Writer) (err error) {
	err = WriteVarInt(writer, h.ProtocolVersion)
	if err != nil {
		return
	}
	err = WriteString(writer, h.ServerAddress)
	if err != nil {
		return
	}
	err = WriteUint16(writer, h.ServerPort)
	if err != nil {
		return
	}
	err = WriteVarInt(writer, h.State)
	if err != nil {
		return
	}
	return
}

func (h *Handshake) Decode(reader io.Reader) error {
	var err error

	h.ProtocolVersion, err = ReadVarInt(reader)
	if err != nil {
		return err
	}

	h.ServerAddress, err = ReadString(reader)
	if err != nil {
		return err
	}
	h.ServerPort, err = ReadUint16(reader)
	if err != nil {
		return err
	}
	h.State, err = ReadVarInt(reader)
	return err
}

func (h *Handshake) ID() int {
	return PACKET_HANDSHAKE
}
