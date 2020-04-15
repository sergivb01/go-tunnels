package packet

import (
	"fmt"
	"io"
)

type LoginStart struct {
	Name string
}

func (s *LoginStart) Encode(writer io.Writer) error {
	if err := WriteString(writer, s.Name); err != nil {
		return fmt.Errorf("error encoding Name: %w", err)
	}
	return nil
}

func (s *LoginStart) Decode(reader io.Reader) error {
	var err error

	s.Name, err = ReadString(reader)
	if err != nil {
		return fmt.Errorf("error reading Name: %w", err)
	}

	return nil
}

func (s *LoginStart) ID() int {
	return HandshakeId
}
