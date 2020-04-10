package proto

import (
	"bufio"
	"log"
	"net"
)

func Start() {
	ln, err := net.Listen("tcp", config.ListenAddress)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Server launched on port", config.ListenAddress)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Print(err)
		} else {
			go handleConnection(conn)
		}
	}
}

func handleConnection(conn net.Conn) {
	log.Printf("%s(#%d) connected.", conn.RemoteAddr().String(), 0)

	player := &Player{
		conn:     conn,
		protocol: V1_10,
		io: &ConnReadWrite{
			rdr: bufio.NewReader(conn),
			wtr: bufio.NewWriter(conn),
		},
		state: HANDSHAKING,
	}

	abc(player)
	abc(player)

	conn.Close()
	log.Printf("%s(#%d) disconnected.", conn.RemoteAddr().String(), 0)
}

func abc(player *Player) {
	pk, err := player.ReadPacket()
	if err != nil {
		log.Printf("error reading packet: %s", err)
		return
	}
	if pk == nil {
		log.Printf("packet is unknown or invalid")
		return
	}
	pk.Handle(player)
	id, proto := pk.Id()
	log.Printf("received packet id=%d and proto=%d\n", id, uint16(proto))
}
