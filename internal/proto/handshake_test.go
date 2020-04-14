package proto

import (
	"io/ioutil"
	"testing"
)

func BenchmarkHandshake_Write(b *testing.B) {
	h := &Handshake{
		ProtocolVersion: 47,
		Address:         "lunar.gg.tunnel.sergitest.dev",
		Port:            25565,
		Status:          1,
	}
	for i := 0; i < b.N * 10; i++ {
		if err := h.Write(ioutil.Discard, "lunar.gg"); err != nil {
			b.Error(err)
		}
	}
}
func BenchmarkHandshake_ReverseWrite(b *testing.B) {
	h := &Handshake{
		ProtocolVersion: 47,
		Address:         "lunar.gg.tunnel.sergitest.dev",
		Port:            25565,
		Status:          1,
	}
	for i := 0; i < b.N * 10; i++ {
		if err := h.ReverseWrite(ioutil.Discard, "lunar.gg"); err != nil {
			b.Error(err)
		}
	}
}
