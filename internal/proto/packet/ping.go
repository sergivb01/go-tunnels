package packet

import (
	"fmt"
	"io"
)

// Ping specifies the https://wiki.vg/Protocol#Ping
type Ping struct {
	Payload int64
}

// Encode encodes the Ping
func (p *Ping) Encode(w io.Writer) error {
	if err := WriteInt64(w, p.Payload); err != nil {
		return fmt.Errorf("error encoding PingPayload: %w", err)
	}
	return nil
}

// Decode decodes a Ping
func (p *Ping) Decode(r io.Reader) error {
	var err error

	p.Payload, err = ReadInt64(r)
	if err != nil {
		return fmt.Errorf("error reading Payload: %w", err)
	}

	return nil
}

// ID returns the Ping-PacketID
func (p *Ping) ID() int {
	return PingID
}
