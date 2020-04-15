package proto

import (
	"io/ioutil"
	"testing"

	"github.com/sergivb01/mctunnel/internal/proto/packet"
)

// TODO: add Read method with a const byte slice of an example handshake

func BenchmarkHandshake_Write(b *testing.B) {
	h := &packet.Handshake{
		ProtocolVersion: 47,
		ServerAddress:   "lunar.gg.tunnel.sergitest.dev",
		ServerPort:      25565,
		State:           1,
	}
	p := NewPacketCodec()
	for i := 0; i < b.N; i++ {
		if err := p.WritePacket(ioutil.Discard, h); err != nil {
			b.Error(err)
		}
	}
}
