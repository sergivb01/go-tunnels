package packet

import (
	"fmt"
	"io"
)

// LoginStart specifies the https://wiki.vg/Protocol#Login_Start
type LoginStart struct {
	Name string
}

// Encode encodes the LoginStart
func (s *LoginStart) Encode(w io.Writer) error {
	if err := WriteString(w, s.Name); err != nil {
		return fmt.Errorf("error encoding Json: %w", err)
	}
	return nil
}

// Decode decodes a LoginStart
func (s *LoginStart) Decode(r io.Reader) error {
	var err error

	s.Name, err = ReadString(r)
	if err != nil {
		return fmt.Errorf("error reading Json: %w", err)
	}

	return nil
}

// ID returns the LoginStart-PacketID
func (s *LoginStart) ID() int {
	return HandshakeID
}
