package packet

import (
	"fmt"
	"io"
)

// ServerStatus specifies the https://wiki.vg/Protocol#Status
type ServerStatus struct {
	ServerName string
	Protocol   int
	Motd       string
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

// Encode encodes the ServerStatus
func (s *ServerStatus) Encode(w io.Writer) error {
	if err := WriteString(w, fmt.Sprintf(baseFormat, s.ServerName, s.Protocol, s.Motd, s.Favicon)); err != nil {
		return fmt.Errorf("error encoding Json: %w", err)
	}
	return nil
}

// Decode should not be used
func (s *ServerStatus) Decode(_ io.Reader) error {
	return fmt.Errorf("not implemented")
}

// ID returns the ServerStatus-PacketID
func (s *ServerStatus) ID() int {
	return HandshakeID
}
