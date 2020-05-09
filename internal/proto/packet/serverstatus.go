package packet

import (
	"fmt"
	"io"
)

type ServerStatus struct {
	ServerName string
	Protocol   int
	MOTD       string
	Favicon    string
}

const baseFormat = `{
    "version": {
        "name": "%s",
        "protocol": %d
    },
    "players": {
        "max": 0,
        "online": 0
    },
    "description": {
        "text": "%s"
    },
    "favicon": "%s"
}`

func (s *ServerStatus) Encode(w io.Writer) error {
	if err := WriteString(w, fmt.Sprintf(baseFormat, s.ServerName, s.Protocol, s.MOTD, s.Favicon)); err != nil {
		return fmt.Errorf("error encoding Json: %w", err)
	}
	return nil
}

func (s *ServerStatus) Decode(r io.Reader) error {
	return fmt.Errorf("not implemented")
}

func (s *ServerStatus) ID() int {
	return HandshakeId
}
