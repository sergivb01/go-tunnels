package proto

import (
	"bytes"
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

func TestEncodeDecode(t *testing.T){
	p := NewPacketCodec()

	h := &packet.Handshake{
		ProtocolVersion: 47,
		ServerAddress:   "lunar.gg.tunnel.sergitest.dev",
		ServerPort:      25565,
		State:           1,
	}
	buff := &bytes.Buffer{}
	if err := p.WritePacket(buff, h); err != nil {
		t.Errorf("error writing packet: %s", err)
	}

	packetID, err := p.ReadPacket(buff)
	if err != nil {
		t.Errorf("error reading packet: %s", err)
	}

	if packetID != packet.HandshakeId {
		t.Errorf("did not receive handshake, instead %d", packetID)
	}

	newH := &packet.Handshake{}
	if err := newH.Decode(buff); err != nil {
		t.Errorf("error decoding : %s", err)
	}

}