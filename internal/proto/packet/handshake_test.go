package packet

import (
	"io/ioutil"
	"testing"

	"github.com/sergivb01/mctunnel/internal/proto"
)

func BenchmarkHandshake_Write(b *testing.B) {
	h := &Handshake{
		ProtocolVersion: 47,
		ServerAddress:   "lunar.gg.tunnel.sergitest.dev",
		ServerPort:      25565,
		State:           1,
	}
	p := proto.NewPacketDecoder()
	for i := 0; i < b.N; i++ {
		if err := p.EncodePacket(ioutil.Discard, h); err != nil {
			b.Error(err)
		}
	}
}
