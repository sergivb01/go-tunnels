package proto

import (
	"log"
	"reflect"
)

var (
	packets = make(map[int64]reflect.Type)
)

type Packet interface {
	Write(*Player) error
	Read(*Player, int) error
	Handle(*Player)
	Id() (int, Protocol)
}

func PacketTypeHash(state State, id int) int64 {
	return int64(id) ^ (int64(state) << 32)
}

func init() {
	packets[PacketTypeHash(HANDSHAKING, 0x00)] = reflect.TypeOf((*PacketHandshake)(nil)).Elem()
	packets[PacketTypeHash(STATUS, 0x00)] = reflect.TypeOf((*PacketStatusRequest)(nil)).Elem()
}

func (player *Player) HandlePacket(id int, length int) (packet Packet, err error) {
	typ := packets[PacketTypeHash(player.state, id)]

	if typ == nil {
		log.Printf("%d -> Unknown packet #%d (state=%d, id=%d)\n", 0, id, player.state, id)

		var buff []byte
		nbr := 0
		if length > 500 {
			buff = make([]byte, 500)
		} else {
			buff = make([]byte, length)
		}

		for nbr < length {
			if length-nbr > 500 {
				player.io.rdr.Read(buff)
				nbr += 500
			} else {
				player.io.rdr.Read(buff[:length-nbr])
				nbr = length
			}
		}
		return nil, nil
	}

	packet, _ = reflect.New(typ).Interface().(Packet)
	if err = packet.Read(player, length); err != nil {
		return nil, err
	}
	return
}
