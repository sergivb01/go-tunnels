package protocol

import (
	"reflect"

	"github.com/sergivb01/mctunnel/internal/protocol/packet"
)

var (
	Packets = map[packet.Direction]map[State]map[int]reflect.Type{
		packet.DirectionServerbound: {
			HandshakeState: {
				0x00: reflect.TypeOf(packet.Handshake{}),
			},
			StatusState: {
				0x00: reflect.TypeOf(packet.StatusRequest{}),
				0x01: reflect.TypeOf(packet.StatusPing{}),
			},
		},
		packet.DirectionClientbound: {
			StatusState: {
				0x00: reflect.TypeOf(packet.StatusResponse{}),
				0x01: reflect.TypeOf(packet.StatusPong{}),
			},
		},
	}
)

func GetPacket(d packet.Direction, s State, id int) (reflect.Type, bool) {
	typ, ok := Packets[d][s][id]
	return typ, ok
}
