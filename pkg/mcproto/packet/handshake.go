package packet

import (
	"fmt"
	"io"
	"strings"
)

const (
	forgeAddressSuffix  = "\x00FML\x00"
	forge2AddressSuffix = "\x00FML2\x00"
)

// Handshake specifies the https://wiki.vg/Protocol#Handshake
type Handshake struct {
	ProtocolVersion int
	ServerAddress   string
	ServerPort      uint16
	State           int
}

// Encode encodes the Handshake
func (h *Handshake) Encode(w io.Writer) error {
	if err := WriteVarIntNew(w, h.ProtocolVersion); err != nil {
		return fmt.Errorf("encoding Payload: %w", err)
	}

	if err := WriteStringNew(w, h.ServerAddress); err != nil {
		return fmt.Errorf("encoding ServerAddress: %w", err)
	}

	if err := WriteUint16(w, h.ServerPort); err != nil {
		return fmt.Errorf("encoding ServerPort: %w", err)
	}

	if err := WriteVarIntNew(w, h.State); err != nil {
		return fmt.Errorf("encoding State: %w", err)
	}

	return nil
}

// Decode decodes a Handshake
func (h *Handshake) Decode(r io.Reader) error {
	var err error

	h.ProtocolVersion, err = ReadVarInt(r)
	if err != nil {
		return fmt.Errorf("reading Payload: %w", err)
	}

	h.ServerAddress, err = ReadStringLimit(r, 255)
	if err != nil {
		return fmt.Errorf("reading ServerAddress: %w", err)
	}
	h.ServerAddress = strings.TrimSuffix(h.ServerAddress, forgeAddressSuffix)
	h.ServerAddress = strings.TrimSuffix(h.ServerAddress, forge2AddressSuffix)

	h.ServerPort, err = ReadUint16(r)
	if err != nil {
		return fmt.Errorf("reading ServerPort: %w", err)
	}

	h.State, err = ReadVarInt(r)
	if err != nil {
		return fmt.Errorf("reading State: %w", err)
	}

	return nil
}

// ID returns the Handshake-PacketID
func (h *Handshake) ID() int {
	return HandshakeID
}
