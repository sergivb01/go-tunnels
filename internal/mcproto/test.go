package mcproto

import (
	"bytes"
)

func (c *connectorImpl) EncodeDecode() {
	h := &Handshake{
		ProtocolVersion: 47,
		ServerAddress:   "lunar.gg",
		ServerPort:      25565,
		NextState:       2,
	}
	c.log.Info().Interface("handshake", h).Msg("original")

	b, err := h.EncodePacket("lunar.gg")
	if err != nil {
		c.log.Error().Err(err).Msg("error encoding custom packet")
		return
	}

	packet, err := ReadPacket(bytes.NewReader(b), 0)
	if err != nil {
		c.log.Error().Err(err).Msg("failed to read packet")
		return
	}
	c.log.Debug().Int("packetID", packet.PacketID).Msg("read packet")

	hand, err := ReadHandshake(packet.Data)
	if err != nil {
		c.log.Error().Err(err).Msg("failed to read handshake")
		return
	}
	c.log.Info().Interface("handshake", hand).Msg("read")
}
