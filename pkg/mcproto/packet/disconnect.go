package packet

import (
	"fmt"
	"io"
)

// LoginDisconnect specifies the https://wiki.vg/Protocol#Disconnect_.28login.29
type LoginDisconnect struct {
	Reason string
}

// Encode encodes the LoginDisconnect
func (d *LoginDisconnect) Encode(w io.Writer) error {
	if err := WriteString(w, d.Reason); err != nil {
		return fmt.Errorf("encoding Reason: %w", err)
	}
	return nil
}

// Decode should not be used
func (d *LoginDisconnect) Decode(_ io.Reader) error {
	return errNotImplemented
}

// ID returns the LoginDisconnect-PacketID
func (d *LoginDisconnect) ID() int {
	return HandshakeID
}
