package proto

import (
	"io/ioutil"
	"testing"
)

func BenchmarkHandshake_Write(t *testing.B) {
	b := &Handshake{
		ProtocolVersion: 47,
		Address:         "lunar.gg.tunnel.sergitest.dev",
		Port:            25565,
		Status:          1,
	}
	for i := 0; i < t.N; i++ {
		if err := b.Write(ioutil.Discard, "lunar.gg"); err != nil {
			t.Error(err)
		}
	}
}
