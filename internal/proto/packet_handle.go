package proto

import (
	"fmt"
	"log"
)

type PacketHandshake struct {
	Protocol Protocol
	Address  string
	Port     uint16
	State    State
}

func (packet *PacketHandshake) Read(player *Player, length int) (err error) {
	protocol, err := player.ReadVarInt()
	if err != nil {
		log.Print(err)
		return
	}
	packet.Protocol = Protocol(protocol)
	packet.Address, err = player.ReadStringLimited(config.BufferConfig.HandshakeAddress)
	if err != nil {
		log.Print(err)
		return
	}
	packet.Port, err = player.ReadUInt16()
	if err != nil {
		log.Print(err)
		return
	}
	state, err := player.ReadVarInt()
	if err != nil {
		log.Print(err)
		return
	}
	packet.State = State(state)
	return
}
func (packet *PacketHandshake) Write(player *Player) (err error) {
	if err := player.WriteVarInt(int(packet.Protocol)); err != nil {
		return fmt.Errorf("error writing protocol handshake: %w", err)
	}

	if err := player.WriteStringRestricted(packet.Address, config.BufferConfig.HandshakeAddress); err != nil {
		return fmt.Errorf("error writing address handshake: %w", err)
	}

	if err := player.WriteUInt16(packet.Port); err != nil {
		return fmt.Errorf("error writing port handshake: %w", err)
	}

	if err := player.WriteVarInt(int(packet.State)); err != nil {
		return fmt.Errorf("error writing state handshake: %w", err)
	}

	return
}
func (packet *PacketHandshake) Handle(player *Player) {
	player.state = packet.State
	player.protocol = packet.Protocol
	player.address = packet.Address
}
func (packet *PacketHandshake) Id() (int, Protocol) {
	return 0x00, V1_10
}

type PacketStatusRequest struct{}

func (packet *PacketStatusRequest) Read(player *Player, length int) (err error) {
	return
}
func (packet *PacketStatusRequest) Write(player *Player) (err error) {
	return
}
func (packet *PacketStatusRequest) Handle(player *Player) {
	max_players := config.MaxPlayers
	motd := config.Motd

	response := PacketStatusResponse{
		Response: fmt.Sprintf(`{"version":{"name":"Typhoon","protocol":%d},"players":{"max":%d,"online":%d,"sample":[]},"description":{"text":"%s"},"favicon":"%s","modinfo":{"type":"FML","modList":[]}}`, player.protocol, max_players, 15, JsonEscape(motd), JsonEscape("")),
	}
	if err := player.WritePacket(&response); err != nil {
		log.Printf("error writing packet StatusResponse to player: %s", err)
	}
}
func (packet *PacketStatusRequest) Id() (int, Protocol) {
	return 0x00, V1_10
}

type PacketStatusResponse struct {
	Response string
}

func (packet *PacketStatusResponse) Read(player *Player, length int) error {
	response, err := player.ReadString()
	if err != nil {
		return fmt.Errorf("error reading response statusresponse: %w", err)
	}
	packet.Response = response
	return nil
}
func (packet *PacketStatusResponse) Write(player *Player) (err error) {
	err = player.WriteString(packet.Response)
	if err != nil {
		log.Print(err)
		return
	}
	return
}
func (packet *PacketStatusResponse) Handle(player *Player) {}
func (packet *PacketStatusResponse) Id() (int, Protocol) {
	return 0x00, V1_10
}
