package packet

import (
	"fmt"
	"io"
)

type LoginStart struct {
	Name string
}

func (s *LoginStart) Encode(w io.Writer) error {
	if err := WriteString(w, s.Name); err != nil {
		return fmt.Errorf("error encoding Json: %w", err)
	}
	return nil
}

func (s *LoginStart) Decode(r io.Reader) error {
	var err error

	s.Name, err = ReadString(r)
	if err != nil {
		return fmt.Errorf("error reading Json: %w", err)
	}

	return nil
}

func (s *LoginStart) ID() int {
	return HandshakeId
}
