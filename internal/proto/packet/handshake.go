package packet

import (
	"fmt"
	"io"
)

type Handshake struct {
	ProtocolVersion int
	ServerAddress   string
	ServerPort      uint16
	State           int
}

func (h *Handshake) Encode(w io.Writer) error {
	if err := WriteVarInt(w, h.ProtocolVersion); err != nil {
		return fmt.Errorf("error encoding ProtocolVersion: %w", err)
	}

	if err := WriteString(w, h.ServerAddress); err != nil {
		return fmt.Errorf("error encoding ServerAddress: %w", err)
	}

	if err := WriteUint16(w, h.ServerPort); err != nil {
		return fmt.Errorf("error encoding ServerPort: %w", err)
	}

	if err := WriteVarInt(w, h.State); err != nil {
		return fmt.Errorf("error encoding State: %w", err)
	}

	return nil
}

func (h *Handshake) Decode(r io.Reader) error {
	var err error

	h.ProtocolVersion, err = ReadVarInt(r)
	if err != nil {
		return fmt.Errorf("error reading ProtocolVersion: %w", err)
	}

	h.ServerAddress, err = ReadString(r)
	if err != nil {
		return fmt.Errorf("error reading ServerAddress: %w", err)
	}

	h.ServerPort, err = ReadUint16(r)
	if err != nil {
		return fmt.Errorf("error reading ServerPort: %w", err)
	}

	h.State, err = ReadVarInt(r)
	if err != nil {
		return fmt.Errorf("error reading State: %w", err)
	}

	return nil
}

func (h *Handshake) ID() int {
	return HandshakeId
}
