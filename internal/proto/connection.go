package proto

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

type State int8

const (
	HANDSHAKING State = iota
	STATUS
	LOGIN
)

type Protocol uint16

const (
	V1_7_2  Protocol = 4
	V1_7_6  Protocol = 5
	V1_8    Protocol = 47
	V1_9    Protocol = 107
	V1_9_1  Protocol = 108
	V1_9_2  Protocol = 109
	V1_9_3  Protocol = 110
	V1_10   Protocol = 210
	V1_11   Protocol = 315
	V1_11_1 Protocol = 316
	V1_12   Protocol = 335
	V1_12_1 Protocol = 338
	V1_12_2 Protocol = 340
	V1_13   Protocol = 393
	V1_13_1 Protocol = 401
	V1_13_2 Protocol = 404
	V1_14   Protocol = 477
	V1_14_1 Protocol = 480
	V1_14_2 Protocol = 485
	V1_14_3 Protocol = 490
	V1_14_4 Protocol = 498
	V1_15   Protocol = 573
	V1_15_1 Protocol = 575
	V1_15_2 Protocol = 578
)

var (
	COMPATIBLE_PROTO = []Protocol{
		V1_7_2, V1_7_6,
		V1_8,
		V1_9, V1_9_1, V1_9_2, V1_9_3,
		V1_10,
		V1_11, V1_11_1,
		V1_12, V1_12_1,
		V1_12_2,
		V1_13,
		V1_13_1,
		V1_13_2,
		V1_14,
		V1_14_1,
		V1_14_2,
		V1_14_3,
		V1_14_4,
		V1_15,
		V1_15_1,
		V1_15_2,
	}
)

func IsCompatible(proto Protocol) bool {
	for _, x := range COMPATIBLE_PROTO {
		if x == proto {
			return true
		}
	}
	return false
}

type Player struct {
	conn     net.Conn
	io       *ConnReadWrite
	protocol Protocol
	state    State
	address  string
}

func NewPlayer(conn net.Conn) *Player {
	return &Player{
		conn:     conn,
		protocol: V1_10,
		io: &ConnReadWrite{
			rdr: bufio.NewReader(conn),
			wtr: bufio.NewWriter(conn),
		},
		state: HANDSHAKING,
	}
}

func (player *Player) ReadPacket() (Packet, error) {
	length, err := player.ReadVarInt()
	if err != nil {
		return nil, err
	}

	id, err := player.ReadVarInt()
	if err != nil {
		return nil, err
	}

	packet, err := player.HandlePacket(id, length)
	if err != nil {
		return nil, err
	}

	if packet != nil {
		log.Printf("#%d -> %d %s", 0, id, fmt.Sprint(packet))
		return packet, nil
	}

	return nil, nil
}

func (player *Player) WritePacket(packet Packet) (err error) {
	buff := newVarBuffer(256)
	tmp := player.io
	player.io = &ConnReadWrite{
		rdr: tmp.rdr,
		wtr: buff,
	}

	id, _ := packet.Id()
	if id == -1 {
		return
	}
	player.WriteVarInt(id)
	packet.Write(player)

	ln := newVarBuffer(0)
	player.io.wtr = ln
	player.WriteVarInt(buff.Len())
	player.io = tmp
	player.conn.Write(ln.Bytes())
	player.conn.Write(buff.Bytes())

	log.Printf("#%d <- %d %s", 0, id, fmt.Sprint(packet))
	return nil
}
